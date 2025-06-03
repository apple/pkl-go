// ===----------------------------------------------------------------------===//
// Copyright © 2024-2025 Apple Inc. and the Pkl project authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   https://www.apache.org/licenses/LICENSE-2.0
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
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/apple/pkl-go/pkl/internal"
	"github.com/apple/pkl-go/pkl/internal/msgapi"
	"github.com/vmihailenco/msgpack/v5"
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
			in:          make(chan msgapi.IncomingMessage),
			out:         make(chan msgapi.OutgoingMessage),
			closed:      make(chan error),
			pklCommand:  pklCommand,
			processDone: make(chan struct{}),
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
	// exited is a flag that indicates evaluator was closed explicitly
	exited      atomicBool
	version     *semver
	pklCommand  []string
	processDone chan struct{}
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

var pklVersionRegex = regexp.MustCompile(fmt.Sprintf("Pkl (%s).*", semverPattern.String()))

func (e *execEvaluator) getVersion() (*semver, error) {
	if e.version != nil {
		return e.version, nil
	}
	cmd, args := e.getCommandAndArgStrings()
	command := exec.Command(cmd, append(args, "--version")...)
	versionCmdOut, err := command.Output()
	if err != nil {
		return nil, err
	}
	matches := pklVersionRegex.FindStringSubmatch(string(versionCmdOut))
	if len(matches) < 2 {
		return nil, fmt.Errorf("failed to get version information from Pkl. Ran `%s`, and got stdout \"%s\"", strings.Join(command.Args, " "), versionCmdOut)
	}
	version, err := parseSemver(matches[1])
	if err != nil {
		return nil, err
	}
	e.version = version
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
		parts := strings.Fields(pklExecEnv)
		return parts[0], parts[1:]
	}
	return "pkl", []string{}
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
	defer close(e.processDone)
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

func (e *execEvaluator) handleSendMessages(stdin io.WriteCloser) {
	defer stdin.Close()

	for msg := range e.out {
		internal.Debug("Sending message: %#v", msg)
		b, err := msg.ToMsgPack()
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
	if e.cmd == nil {
		return nil
	}
	e.exited.set(true)
	close(e.in)
	close(e.out)
	close(e.closed)

	return e.enforceKillOnTimeout()
}

func (e *execEvaluator) enforceKillOnTimeout() error {
	select {
	case <-time.After(5 * time.Second):
		if err := killProcess(e.cmd.Process); err != nil {
			return fmt.Errorf("failed to kill process %d: %v", e.cmd.Process.Pid, err)
		}
	case <-e.processDone:
		// The process has finished
	}
	return nil
}
