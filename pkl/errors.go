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
	"errors"
	"fmt"
)

// EvalError is an error that occurs during the normal evaluation of Pkl code.
//
// This means that Pkl evaluation occurred, and the Pkl runtime produced an error.
type EvalError struct {
	ErrorOutput string
}

var _ error = (*EvalError)(nil)

func (r *EvalError) Error() string {
	return r.ErrorOutput
}

// Is implements the interface expected by errors.Is.
func (r *EvalError) Is(err error) bool {
	if err == nil {
		return false
	}
	var evalError *EvalError
	ok := errors.As(err, &evalError)
	return ok
}

// InternalError indicates that an unexpected error occurred.
type InternalError struct {
	err error
}

var _ error = (*InternalError)(nil)

func (r *InternalError) Error() string {
	return fmt.Sprintf("an internal error ocurred: %v", r.err)
}

// Is implements the interface expected by errors.Is.
func (r *InternalError) Is(err error) bool {
	if err == nil {
		return false
	}
	var internalError *InternalError
	ok := errors.As(err, &internalError)
	return ok
}

// Unwrap implements the interface expected by errors.Unwrap.
func (r *InternalError) Unwrap() error {
	return r.err
}
