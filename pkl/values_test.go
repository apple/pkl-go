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
	"testing"
	"time"

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

func TestDataSize_ConvertToUnit(t *testing.T) {
	tests := []struct {
		name     string
		input    DataSize
		expected DataSize
	}{
		{
			name: "bytes_to_gigabytes",
			input: DataSize{
				Value: 1.0,
				Unit:  Bytes,
			},
			expected: DataSize{
				Value: 0.000_000_001,
				Unit:  Gigabytes,
			},
		},
		{
			name: "megabytes_to_bytes",
			input: DataSize{
				Value: 1.0,
				Unit:  Megabytes,
			},
			expected: DataSize{
				Value: 1_000_000,
				Unit:  Bytes,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.input.ToUnit(test.expected.Unit)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestDuration_String(t *testing.T) {
	tests := map[string]struct {
		d    Duration
		want string
	}{
		"should successfully return valid seconds string": {
			d: Duration{
				Unit: Second,
			},
			want: "s",
		},
		"should successfully return valid milliseconds string": {
			d: Duration{
				Unit: Millisecond,
			},
			want: "ms",
		},
		"should successfully return valid nanoseconds string": {
			d: Duration{
				Unit: Nanosecond,
			},
			want: "ns",
		},
		"should successfully return valid microseconds string": {
			d: Duration{
				Unit: Microsecond,
			},
			want: "us",
		},
		"should successfully return valid minute string": {
			d: Duration{
				Unit: Minute,
			},
			want: "min",
		},
		"should successfully return valid hour string": {
			d: Duration{
				Unit: Hour,
			},
			want: "h",
		},
		"should successfully return valid day string": {
			d: Duration{
				Unit: Day,
			},
			want: "d",
		},
		"should return invalid string when provided unknown DurationUnit": {
			d: Duration{
				Unit: DurationUnit(-1),
			},
			want: "<invalid>",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, fmt.Sprintf("%s", tc.d.Unit.String()))
		})
	}
}

func TestDuration_GoDuration(t *testing.T) {
	t.Run("should successfully convert to GoDuration in seconds", func(t *testing.T) {
		t.Parallel()
		d := Duration{
			Value: 1,
			Unit:  Second,
		}
		assert.Equal(t, d.GoDuration(), 1*time.Second)
	})
}
