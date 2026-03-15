//===----------------------------------------------------------------------===//
// Copyright © 2026 Apple Inc. and the Pkl project authors. All rights reserved.
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
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvalError_Error(t *testing.T) {
	tests := map[string]struct {
		err  EvalError
		want string
	}{
		"should return error output message": {
			err:  EvalError{ErrorOutput: "some eval error"},
			want: "some eval error",
		},
		"should return empty string when error output is empty": {
			err:  EvalError{},
			want: "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, tc.err.Error())
		})
	}
}

func TestEvalError_Is(t *testing.T) {
	tests := map[string]struct {
		target error
		want   bool
	}{
		"should return true when target is an EvalError": {
			target: &EvalError{ErrorOutput: "some error"},
			want:   true,
		},
		"should return true when target is a different EvalError": {
			target: &EvalError{ErrorOutput: "different error"},
			want:   true,
		},
		"should return false when target is nil": {
			target: nil,
			want:   false,
		},
		"should return false when target is a different error type": {
			target: fmt.Errorf("some other error"),
			want:   false,
		},
		"should return false when target is an InternalError": {
			target: &InternalError{err: fmt.Errorf("internal")},
			want:   false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			evalErr := &EvalError{ErrorOutput: "eval error"}
			assert.Equal(t, tc.want, evalErr.Is(tc.target))
		})
	}
}

func TestInternalError_Error(t *testing.T) {
	tests := map[string]struct {
		err  InternalError
		want string
	}{
		"should return formatted internal error message": {
			err:  InternalError{err: fmt.Errorf("something broke")},
			want: "an internal error occurred: something broke",
		},
		"should return formatted message with nil inner error": {
			err:  InternalError{err: nil},
			want: "an internal error occurred: <nil>",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, tc.err.Error())
		})
	}
}

func TestInternalError_Is(t *testing.T) {
	tests := map[string]struct {
		target error
		want   bool
	}{
		"should return true when target is an InternalError": {
			target: &InternalError{err: fmt.Errorf("some error")},
			want:   true,
		},
		"should return true when target is a different InternalError": {
			target: &InternalError{err: fmt.Errorf("different error")},
			want:   true,
		},
		"should return false when target is nil": {
			target: nil,
			want:   false,
		},
		"should return false when target is a different error type": {
			target: fmt.Errorf("some other error"),
			want:   false,
		},
		"should return false when target is an EvalError": {
			target: &EvalError{ErrorOutput: "eval"},
			want:   false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			internalErr := &InternalError{err: fmt.Errorf("internal error")}
			assert.Equal(t, tc.want, internalErr.Is(tc.target))
		})
	}
}

func TestInternalError_Unwrap(t *testing.T) {
	tests := map[string]struct {
		inner error
		want  error
	}{
		"should return the wrapped error": {
			inner: fmt.Errorf("wrapped error"),
			want:  fmt.Errorf("wrapped error"),
		},
		"should return nil when inner error is nil": {
			inner: nil,
			want:  nil,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			internalErr := &InternalError{err: tc.inner}
			assert.Equal(t, tc.want, internalErr.Unwrap())
		})
	}
}

func TestEvalError_WorksWithErrorsIs(t *testing.T) {
	t.Parallel()
	evalErr := &EvalError{ErrorOutput: "eval error"}
	wrappedErr := fmt.Errorf("wrapped: %w", evalErr)
	assert.True(t, errors.Is(wrappedErr, &EvalError{}))
}

func TestInternalError_WorksWithErrorsIs(t *testing.T) {
	t.Parallel()
	internalErr := &InternalError{err: fmt.Errorf("something broke")}
	wrappedErr := fmt.Errorf("wrapped: %w", internalErr)
	assert.True(t, errors.Is(wrappedErr, &InternalError{}))
}
