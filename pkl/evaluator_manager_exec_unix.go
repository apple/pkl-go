//===----------------------------------------------------------------------===//
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
//===----------------------------------------------------------------------===//

//go:build unix

package pkl

import (
	"os"
	"os/exec"
	"syscall"
)

func (e *execEvaluator) getStartCommand() *exec.Cmd {
	exe, arg := e.getCommandAndArgStrings()
	cmd := exec.Command(exe, append(arg, "server")...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd
}

// killProcess kills the process's entire group
func killProcess(proc *os.Process) error {
	pgid, err := syscall.Getpgid(proc.Pid)
	if err != nil {
		return err
	}
	// negative pid indicates to send the signal to the whole pg
	return syscall.Kill(-pgid, syscall.SIGKILL)
}
