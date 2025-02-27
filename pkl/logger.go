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

package pkl

import (
	"fmt"
	"io"
	"os"
)

// Logger is the interface for logging messages emitted by the Pkl evaluator.
//
// To set a logger, register it on EvaluatorOptions.Logger when building an Evaluator.
type Logger interface {
	// Trace logs the given message on level TRACE.
	Trace(message string, frameUri string)

	// Warn logs the given message on level WARN.
	Warn(message string, frameUri string)
}

// NewLogger builds a logger that writes to the provided output stream,
// using the default formatting.
func NewLogger(out io.Writer) Logger {
	return &logger{out}
}

// FormatLogMessage returns the default formatter for log messages.
func FormatLogMessage(level, message, frameUri string) string {
	return fmt.Sprintf("pkl: %s: %s (%s)\n", level, message, frameUri)
}

type logger struct {
	out io.Writer
}

func (s logger) Trace(message string, frameUri string) {
	_, _ = s.out.Write([]byte(FormatLogMessage("TRACE", message, frameUri)))
}

func (s logger) Warn(message string, frameUri string) {
	_, _ = s.out.Write([]byte(FormatLogMessage("WARN", message, frameUri)))
}

var _ Logger = (*logger)(nil)

// StderrLogger is a logger that writes to standard error.
//
//goland:noinspection GoUnusedGlobalVariable
var StderrLogger = NewLogger(os.Stdout)

// NoopLogger is a logger that discards all messages.
var NoopLogger = NewLogger(io.Discard)
