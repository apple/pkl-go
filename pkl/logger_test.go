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
	"io"
	"testing"
)

type writerMock struct {
	err   error
	bytes int
}

func (w writerMock) Write(_ []byte) (n int, err error) {
	return w.bytes, w.err
}

func TestLogger(t *testing.T) {
	tests := map[string]struct {
		writerMock io.Writer
		msg        string
		frameURI   string
	}{
		"should successfully log a message as trace and warn": {
			writerMock: writerMock{err: nil, bytes: 20},
			msg:        "test message",
			frameURI:   "test",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			lgr := NewLogger(tc.writerMock)
			lgr.Trace(tc.msg, tc.frameURI)
			lgr.Warn(tc.msg, tc.frameURI)
		})
	}
}
