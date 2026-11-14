//===----------------------------------------------------------------------===//
// Copyright © 2024-2026 Apple Inc. and the Pkl project authors. All rights reserved.
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

func TestDuration_GoDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    Duration
		expected time.Duration
	}{
		{
			name:     "whole_seconds",
			input:    Duration{Value: 5, Unit: Second},
			expected: 5 * time.Second,
		},
		{
			name:     "fractional_minutes",
			input:    Duration{Value: 1.5, Unit: Minute},
			expected: 90 * time.Second,
		},
		{
			name:     "fractional_seconds",
			input:    Duration{Value: 1.005, Unit: Second},
			expected: 1005 * time.Millisecond,
		},
		{
			name:     "fractional_milliseconds",
			input:    Duration{Value: 33.3, Unit: Millisecond},
			expected: 33300 * time.Microsecond,
		},
		{
			name:     "fractional_microseconds",
			input:    Duration{Value: 1.001, Unit: Microsecond},
			expected: 1001 * time.Nanosecond,
		},
		{
			name:     "negative_fractional_seconds",
			input:    Duration{Value: -1.005, Unit: Second},
			expected: -1005 * time.Millisecond,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.input.GoDuration())
		})
	}
}
