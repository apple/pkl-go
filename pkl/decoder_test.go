//===----------------------------------------------------------------------===//
// Copyright © 2026-2027 Apple Inc. and the Pkl project authors. All rights reserved.
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
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"
)

func TestDecoder_Decode(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		typ         reflect.Type
		data        func(t *testing.T) []byte
		schemas     map[string]reflect.Type
		want        reflect.Value
		expectedErr error
	}{
		"should successfully decode a primitive string type": {
			typ: reflect.TypeOf(""),
			data: func(t *testing.T) []byte {
				var buf bytes.Buffer
				enc := msgpack.NewEncoder(&buf)
				assert.NoError(t, enc.EncodeString(""))
				return buf.Bytes()
			},
			want:        reflect.ValueOf(""),
			expectedErr: nil,
		},
		"should successfully decode a primitive bool type": {
			typ: reflect.TypeOf(true),
			data: func(t *testing.T) []byte {
				var buf bytes.Buffer
				enc := msgpack.NewEncoder(&buf)
				assert.NoError(t, enc.EncodeBool(true))
				return buf.Bytes()
			},
			want:        reflect.ValueOf(true),
			expectedErr: nil,
		},
		"should successfully decode a primitive int type": {
			typ: reflect.TypeOf(0),
			data: func(t *testing.T) []byte {
				var buf bytes.Buffer
				enc := msgpack.NewEncoder(&buf)
				assert.NoError(t, enc.EncodeInt(0))
				return buf.Bytes()
			},
			want:        reflect.ValueOf(0),
			expectedErr: nil,
		},
		"should successfully decode a primitive int8 type": {
			typ: reflect.TypeOf(int8(0)),
			data: func(t *testing.T) []byte {
				var buf bytes.Buffer
				enc := msgpack.NewEncoder(&buf)
				assert.NoError(t, enc.EncodeInt8(int8(0)))
				return buf.Bytes()
			},
			want:        reflect.ValueOf(int8(0)),
			expectedErr: nil,
		},
		"should successfully decode a primitive int16 type": {
			typ: reflect.TypeOf(int16(0)),
			data: func(t *testing.T) []byte {
				var buf bytes.Buffer
				enc := msgpack.NewEncoder(&buf)
				assert.NoError(t, enc.EncodeInt16(int16(0)))
				return buf.Bytes()
			},
			want:        reflect.ValueOf(int16(0)),
			expectedErr: nil,
		},
		"should successfully decode a primitive int32 type": {
			typ: reflect.TypeOf(int32(0)),
			data: func(t *testing.T) []byte {
				var buf bytes.Buffer
				enc := msgpack.NewEncoder(&buf)
				assert.NoError(t, enc.EncodeInt32(int32(0)))
				return buf.Bytes()
			},
			want:        reflect.ValueOf(int32(0)),
			expectedErr: nil,
		},
		"should successfully decode a primitive int64 type": {
			typ: reflect.TypeOf(int64(0)),
			data: func(t *testing.T) []byte {
				var buf bytes.Buffer
				enc := msgpack.NewEncoder(&buf)
				assert.NoError(t, enc.EncodeInt64(int64(0)))
				return buf.Bytes()
			},
			want:        reflect.ValueOf(int64(0)),
			expectedErr: nil,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var data []byte
			// load custom data
			if tc.data != nil {
				data = tc.data(t)
			}
			d := newDecoder(data, tc.schemas)

			got, err := d.Decode(tc.typ)
			if tc.expectedErr != nil {
				assert.Nil(t, got)
				assert.EqualError(t, err, tc.expectedErr.Error())
				return
			}
			assert.Nil(t, err)

			assert.NotNil(t, got)
			assert.NotNil(t, tc.want)

			assert.Equal(t, tc.want.Interface(), got.Interface())
		})
	}

}
