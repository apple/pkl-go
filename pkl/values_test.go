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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataSize_String(t *testing.T) {
	tests := []struct {
		name     string
		input    DataSize
		expected string
	}{
		{
			name: "bytes",
			input: DataSize{
				Value: 1.0,
				Unit:  Bytes,
			},
			expected: "1.b",
		},
		{
			name: "kebibytes",
			input: DataSize{
				Value: 5.3,
				Unit:  Kibibytes,
			},
			expected: "5.3.kib",
		},
		{
			name: "invalid",
			input: DataSize{
				Value: 5.0,
			},
			expected: "5.<invalid>",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.input.String())
		})
	}
}
