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
			t.Parallel()
			assert.Equal(t, test.expected, test.input.String())
		})
	}
}

func TestDataSizeUnit_String(t *testing.T) {
	tests := []struct {
		name string
		unit DataSizeUnit
		want string
	}{
		{name: "byte", unit: Bytes, want: "b"},
		{name: "kilobyte", unit: Kilobytes, want: "kb"},
		{name: "kibibyte", unit: Kibibytes, want: "kib"},
		{name: "megabyte", unit: Megabytes, want: "mb"},
		{name: "mebibyte", unit: Mebibytes, want: "mib"},
		{name: "gigabyte", unit: Gigabytes, want: "gb"},
		{name: "gibibyte", unit: Gibibytes, want: "gib"},
		{name: "terabyte", unit: Terabytes, want: "tb"},
		{name: "tebibyte", unit: Tebibytes, want: "tib"},
		{name: "petabyte", unit: Petabytes, want: "pb"},
		{name: "pebibyte", unit: Pebibytes, want: "pib"},
		{name: "invalid", unit: DataSizeUnit(-1), want: "<invalid>"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.want, test.unit.String())
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
		"should successfully return valid second string": {
			d: Duration{
				Unit: Second,
			},
			want: "s",
		},
		"should successfully return valid millisecond string": {
			d: Duration{
				Unit: Millisecond,
			},
			want: "ms",
		},
		"should successfully return valid nanosecond string": {
			d: Duration{
				Unit: Nanosecond,
			},
			want: "ns",
		},
		"should successfully return valid microsecond string": {
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
	t.Run("should successfully convert to GoDuration in second", func(t *testing.T) {
		t.Parallel()
		d := Duration{
			Value: 1,
			Unit:  Second,
		}
		assert.Equal(t, d.GoDuration(), 1*time.Second)
	})
}

func TestDuration_UnmarshalBinary(t *testing.T) {
	tests := map[string]struct {
		input       []byte
		want        DurationUnit
		expectedErr error
	}{
		"should successfully unmarshal binary from nanosecond": {
			input: []byte("ns"),
			want:  Nanosecond,
		},
		"should successfully unmarshal binary from  microsecond": {
			input: []byte("us"),
			want:  Microsecond,
		},
		"should successfully unmarshal binary from  millisecond": {
			input: []byte("ms"),
			want:  Millisecond,
		},
		"should successfully unmarshal binary from  second": {
			input: []byte("s"),
			want:  Second,
		},
		"should successfully unmarshal binary from  minute": {
			input: []byte("min"),
			want:  Minute,
		},
		"should successfully unmarshal binary from  hour": {
			input: []byte("h"),
			want:  Hour,
		},
		"should successfully unmarshal binary from  day": {
			input: []byte("d"),
			want:  Day,
		},
		"should fail unmarshal binary from unknown field": {
			input:       []byte("unknown"),
			expectedErr: fmt.Errorf("unrecognized Duration unit: `unknown`"),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var unit DurationUnit
			err := unit.UnmarshalBinary(tc.input)

			assert.Equal(t, tc.want, unit)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestDataSizeUnit_UnmarshalBinary(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		want        DataSizeUnit
		expectedErr error
	}{
		{name: "byte", input: []byte("b"), want: Bytes},
		{name: "kilobyte", input: []byte("kb"), want: Kilobytes},
		{name: "kibibyte", input: []byte("kib"), want: Kibibytes},
		{name: "megabyte", input: []byte("mb"), want: Megabytes},
		{name: "mebibyte", input: []byte("mib"), want: Mebibytes},
		{name: "gigabyte", input: []byte("gb"), want: Gigabytes},
		{name: "gibibyte", input: []byte("gib"), want: Gibibytes},
		{name: "terabyte", input: []byte("tb"), want: Terabytes},
		{name: "tebibyte", input: []byte("tib"), want: Tebibytes},
		{name: "petabyte", input: []byte("pb"), want: Petabytes},
		{name: "pebibyte", input: []byte("pib"), want: Pebibytes},
		{name: "unknown", input: []byte("unknown"), expectedErr: fmt.Errorf("unrecognized DataSize unit: `unknown`")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var unit DataSizeUnit
			err := unit.UnmarshalBinary(test.input)

			assert.Equal(t, test.want, unit)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}
