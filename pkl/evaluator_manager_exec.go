// ===----------------------------------------------------------------------===//
// Copyright © 2024 Apple Inc. and the Pkl project authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ===----------------------------------------------------------------------===//
package pkl

import (
	"fmt"
	"github.com/apple/pkl-go/pkl/internal"
	"github.com/apple/pkl-go/pkl/internal/msgapi"
	"github.com/vmihailenco/msgpack/v5"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

// NewEvaluatorManager creates a new EvaluatorManager.
func NewEvaluatorManager() EvaluatorManager {
	return NewEvaluatorManagerWithCommand(nil)
}

// NewEvaluatorManagerWithCommand creates a new EvaluatorManager using the given pkl command.
//
// The first element in pklCmd is treated as the command to run.
// Any additional elements are treated as arguments to be passed to the process.
// pklCmd is treated as the base command that spawns Pkl.
// For example, the below snippet spawns the command /opt/bin/pkl.
//
//	NewEvaluatorManagerWithCommand([]string{"/opt/bin/pkl"})
func NewEvaluatorManagerWithCommand(pklCommand []string) EvaluatorManager {
	m := &evaluatorManager{
		impl: &execEvaluator{
			in:         make(chan msgapi.IncomingMessage),
			out:        make(chan msgapi.OutgoingMessage),
			closed:     make(chan error),
			pklCommand: pklCommand,
		},
		interrupts:        &sync.Map{},
		evaluators:        &sync.Map{},
		pendingEvaluators: &sync.Map{},
	}
	go m.listen()
	go m.listenForImplClose()
	return m
}

type execEvaluator struct {
	cmd    *exec.Cmd
	in     chan msgapi.IncomingMessage
	out    chan msgapi.OutgoingMessage
	closed chan error
	// exited is a flag that indicates evaluator was closed explicity
	exited     atomicBool
	version    string
	pklCommand []string
}

func (e *execEvaluator) inChan() chan msgapi.IncomingMessage {
	return e.in
}

func (e *execEvaluator) outChan() chan msgapi.OutgoingMessage {
	return e.out
}

func (e *execEvaluator) closedChan() chan error {
	return e.closed
}

var semverPattern = regexp.MustCompile(`(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?`)

var pklVersionRegex = regexp.MustCompile(fmt.Sprintf("Pkl (%s).*", semverPattern.String()))

func (e *execEvaluator) getVersion() (string, error) {
	if e.version != "" {
		return e.version, nil
	}
	cmd, args := e.getCommandAndArgStrings()
	command := exec.Command(cmd, append(args, "--version")...)
	versionCmdOut, err := command.Output()
	if err != nil {
		return "", err
	}
	version := pklVersionRegex.FindStringSubmatch(string(versionCmdOut))
	if len(version) < 2 {
		return "", fmt.Errorf("failed to get version information from Pkl. Ran `%s`, and got stdout \"%s\"", strings.Join(command.Args, " "), versionCmdOut)
	}
	e.version = version[1]
	return e.version, nil
}

var _ evaluatorManagerImpl = (*execEvaluator)(nil)

func (e *execEvaluator) getCommandAndArgStrings() (string, []string) {
	if len(e.pklCommand) > 0 {
		return e.pklCommand[0], e.pklCommand[1:]
	}
	pklExecEnv := os.Getenv("PKL_EXEC")
	if pklExecEnv != "" {
		// this previously required the `--server` argument, and is no longer needed.
		// strip it if exists.
		pklExecEnv = strings.Replace(pklExecEnv, " --server", "", 1)
		parts := strings.Split(pklExecEnv, " ")
		return parts[0], parts[1:]
	}
	return "pkl", []string{}
}

func (e *execEvaluator) getStartCommand() *exec.Cmd {
	cmd, arg := e.getCommandAndArgStrings()
	return exec.Command(cmd, append(arg, "server")...)
}

func (e *execEvaluator) init() error {
	e.cmd = e.getStartCommand()
	e.cmd.Env = os.Environ()
	e.cmd.Stderr = os.Stderr
	stdin, err := e.cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := e.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	go e.readIncomingMessages(stdout)
	go e.handleSendMessages(stdin)
	internal.Debug("Spawning command: %s", e.cmd)
	err = e.cmd.Start()
	if err != nil {
		return err
	}
	go e.listenForProcessClose()
	return nil
}

// listenForProcessClose notifies the evaluator manager when the process quits.
func (e *execEvaluator) listenForProcessClose() {
	err := e.cmd.Wait()
	// e.exited gets set if closed explicitly.
	if e.exited.get() {
		return
	}
	e.closed <- err
}

func (e *execEvaluator) readIncomingMessages(stdout io.Reader) {
	dec := msgpack.NewDecoder(stdout)
	for {
		msg, err := msgapi.Decode(dec)
		if e.exited.get() || err == io.EOF {
			break
		}
		if err != nil {
			e.closed <- &InternalError{err: err}
			return
		}
		internal.Debug("Received message: %#v", msg)
		e.in <- msg
	}
}

func (e *execEvaluator) handleSendMessages(stdin io.Writer) {
	for msg := range e.out {
		internal.Debug("Sending message: %#v", msg)
		b, err := msg.ToMsgPack()
		if e.exited.get() {
			return
		}
		if err != nil {
			e.closed <- &InternalError{err: err}
			return
		}
		if _, err = stdin.Write(b); err != nil {
			if !e.exited.get() {
				e.closed <- &InternalError{err: err}
			}
			return
		}
	}
}

func (e *execEvaluator) deinit() error {
	// `cmd` is nil until an evaluator is initialized through NewEvaluator. If the manager is closed without any
	// evaluators being initialized, `e.cmd` will be nil.
	if e.cmd == nil {
		return nil
	}
	e.exited.set(true)
	close(e.in)
	close(e.out)
	close(e.closed)
	// TODO: graceful shutdown
	if err := e.cmd.Process.Signal(os.Interrupt); err != nil {
		internal.Debug("Failed to interrupt process: %v", err)
		return e.cmd.Process.Kill()
	}
	select {
	case <-time.After(5 * time.Second):
		// If the process does not exit within the timeout, kill it.
		if killErr := e.cmd.Process.Kill(); killErr != nil {
			internal.Debug("Failed to kill process after timeout: %v", killErr)
			return killErr
		}
	case err := <-e.closed:
		if err != nil {
			internal.Debug("Process exited with error: %v", err)
			// Forcefully kill the process if it exited with an error.
			if killErr := e.cmd.Process.Kill(); killErr != nil {
				internal.Debug("Failed to kill process after error: %v", killErr)
				return killErr
			}
		}
	}
	return nil
}
