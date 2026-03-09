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
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"
)

func TestDecoder_Decode(t *testing.T) {
	t.Parallel()
	type dummyStruct struct {
		Name string
	}
	tests := map[string]struct {
		typ         reflect.Type
		data        func(t *testing.T, enc *msgpack.Encoder)
		schemas     map[string]reflect.Type
		want        any
		expectedErr error
	}{
		"should successfully decode a primitive string type": {
			typ: reflect.TypeOf(""),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeString(""))
			},
			want:        "",
			expectedErr: nil,
		},
		"should successfully decode a primitive bool type": {
			typ: reflect.TypeOf(true),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeBool(true))
			},
			want:        true,
			expectedErr: nil,
		},
		"should successfully decode a primitive go duration type": {
			typ: reflect.TypeOf(time.Duration(10)),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeFloat64(10.0))
				assert.NoError(t, enc.EncodeString("ns"))
			},
			want:        Duration{Value: 10.0, Unit: Nanosecond},
			expectedErr: nil,
		},
		"should successfully decode a primitive int type": {
			typ: reflect.TypeOf(0),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeInt(0))
			},
			want:        0,
			expectedErr: nil,
		},
		"should successfully decode a primitive int8 type": {
			typ: reflect.TypeOf(int8(0)),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeInt8(int8(0)))
			},
			want:        int8(0),
			expectedErr: nil,
		},
		"should successfully decode a primitive int16 type": {
			typ: reflect.TypeOf(int16(0)),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeInt16(int16(0)))
			},
			want:        int16(0),
			expectedErr: nil,
		},
		"should successfully decode a primitive int32 type": {
			typ: reflect.TypeOf(int32(0)),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeInt32(int32(0)))
			},
			want:        int32(0),
			expectedErr: nil,
		},
		"should successfully decode a primitive int64 type": {
			typ: reflect.TypeOf(int64(0)),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeInt64(int64(0)))
			},
			want:        int64(0),
			expectedErr: nil,
		},
		"should successfully decode a primitive uint type": {
			typ: reflect.TypeOf(uint(0)),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeUint(uint64(0)))
			},
			want:        uint(0),
			expectedErr: nil,
		},
		"should successfully decode a primitive uint8 type": {
			typ: reflect.TypeOf(uint8(0)),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeUint8(uint8(0)))
			},
			want:        uint8(0),
			expectedErr: nil,
		},
		"should successfully decode a primitive uint16 type": {
			typ: reflect.TypeOf(uint16(0)),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeUint16(uint16(0)))
			},
			want:        uint16(0),
			expectedErr: nil,
		},
		"should successfully decode a primitive uint32 type": {
			typ: reflect.TypeOf(uint32(0)),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeUint32(uint32(0)))
			},
			want:        uint32(0),
			expectedErr: nil,
		},
		"should successfully decode a primitive uint64 type": {
			typ: reflect.TypeOf(uint64(0)),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeUint64(uint64(0)))
			},
			want:        uint64(0),
			expectedErr: nil,
		},
		"should successfully decode a primitive float64 type": {
			typ: reflect.TypeOf(float64(0)),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeFloat64(float64(0)))
			},
			want:        float64(0),
			expectedErr: nil,
		},
		"should successfully decode slices with one element": {
			typ: reflect.TypeOf([]string{"1"}),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeArrayLen(2))
				assert.NoError(t, enc.EncodeInt(codeList))
				assert.NoError(t, enc.EncodeArrayLen(1))
				assert.NoError(t, enc.EncodeString("1"))
			},
			want:        []string{"1"},
			expectedErr: nil,
		},
		"should successfully decode maps": {
			typ: reflect.TypeOf(map[string]bool{}),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeArrayLen(2))
				assert.NoError(t, enc.EncodeInt(codeMap))
				assert.NoError(t, enc.EncodeMapLen(1))
				assert.NoError(t, enc.EncodeString("a"))
				assert.NoError(t, enc.EncodeBool(true))
			},
			want:        map[string]bool{"a": true},
			expectedErr: nil,
		},
		"should successfully decode struct": {
			typ: reflect.TypeOf(dummyStruct{}),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeArrayLen(4))
				assert.NoError(t, enc.EncodeInt(codeObject))
				assert.NoError(t, enc.EncodeString("dummyStruct"))
				assert.NoError(t, enc.EncodeString("my.module.dummyStruct"))
				assert.NoError(t, enc.EncodeArrayLen(1)) // 1 member
				assert.NoError(t, enc.EncodeArrayLen(3)) // for members: code + name + value
				assert.NoError(t, enc.EncodeInt(codeObjectMemberProperty))
				assert.NoError(t, enc.EncodeString("Name"))
				assert.NoError(t, enc.EncodeString("Alice"))
			},
			schemas: map[string]reflect.Type{
				"dummyStruct": reflect.TypeOf(dummyStruct{}),
			},
			want:        dummyStruct{Name: "Alice"},
			expectedErr: nil,
		},
		"should successfully decode an interface from string value": {
			typ: reflect.TypeOf((*interface{})(nil)).Elem(),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeString("hello"))
			},
			want:        "hello",
			expectedErr: nil,
		},
		"should successfully decode a duration type": {
			typ: reflect.TypeOf(Duration{}),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeArrayLen(3))
				assert.NoError(t, enc.EncodeInt(codeDuration))
				assert.NoError(t, enc.EncodeFloat64(10.0))
				assert.NoError(t, enc.EncodeString("s"))
			},
			want:        Duration{Value: 10.0, Unit: Second},
			expectedErr: nil,
		},
		"should successfully decode a nil as nil pointer": {
			typ: reflect.TypeOf((*string)(nil)),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeNil())
			},
			want:        (*string)(nil),
			expectedErr: nil,
		},
		"should successfully decode a pointer type with non-empty value not nil": {
			typ: reflect.TypeOf(""),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeString("hello"))
			},
			want:        "hello",
			expectedErr: nil,
		},
		"should successfully decode float64 from int64 encoding": {
			typ: reflect.TypeOf(float64(0)),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeInt64(int64(42)))
			},
			want:        float64(42),
			expectedErr: nil,
		},
		"should successfully decode a DataSize": {
			typ: reflect.TypeOf(DataSize{}),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeArrayLen(3))
				assert.NoError(t, enc.EncodeInt(codeDataSize))
				assert.NoError(t, enc.EncodeFloat64(512.0))
				assert.NoError(t, enc.EncodeString("mb"))
			},
			want:        DataSize{Value: 512.0, Unit: Megabytes},
			expectedErr: nil,
		},
		"should successfully decode a Pair": {
			typ: reflect.TypeOf(Pair[string, int]{}),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeArrayLen(3))
				assert.NoError(t, enc.EncodeInt(codePair))
				assert.NoError(t, enc.EncodeString("key"))
				assert.NoError(t, enc.EncodeInt(42))
			},
			want:        Pair[string, int]{First: "key", Second: 42},
			expectedErr: nil,
		},
		"should successfully decode a Pair with interface type": {
			typ: reflect.TypeOf((*interface{})(nil)).Elem(),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeArrayLen(3))
				assert.NoError(t, enc.EncodeInt(codePair))
				assert.NoError(t, enc.EncodeString("first"))
				assert.NoError(t, enc.EncodeString("second"))
			},
			want:        Pair[any, any]{First: "first", Second: "second"},
			expectedErr: nil,
		},
		"should successfully decode an IntSeq": {
			typ: reflect.TypeOf(IntSeq{}),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeArrayLen(4))
				assert.NoError(t, enc.EncodeInt(codeIntSeq))
				assert.NoError(t, enc.EncodeInt(1))
				assert.NoError(t, enc.EncodeInt(10))
				assert.NoError(t, enc.EncodeInt(2))
			},
			want:        IntSeq{Start: 1, End: 10, Step: 2},
			expectedErr: nil,
		},
		"should successfully decode a Regex": {
			typ: reflect.TypeOf(Regex{}),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeArrayLen(2))
				assert.NoError(t, enc.EncodeInt(codeRegex))
				assert.NoError(t, enc.EncodeString("[a-z]+"))
			},
			want:        Regex{Pattern: "[a-z]+"},
			expectedErr: nil,
		},
		"should successfully decode a Class": {
			typ: reflect.TypeOf(Class{}),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeArrayLen(3))
				assert.NoError(t, enc.EncodeInt(codeClass))
				assert.NoError(t, enc.EncodeString("MyClass"))
				assert.NoError(t, enc.EncodeString("my.module"))
			},
			want:        Class{Name: "MyClass", ModuleUri: "my.module"},
			expectedErr: nil,
		},
		"should successfully decode a TypeAlias": {
			typ: reflect.TypeOf(TypeAlias{}),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeArrayLen(3))
				assert.NoError(t, enc.EncodeInt(codeTypeAlias))
				assert.NoError(t, enc.EncodeString("MyAlias"))
				assert.NoError(t, enc.EncodeString("my.module"))
			},
			want:        TypeAlias{Name: "MyAlias", ModuleUri: "my.module"},
			expectedErr: nil,
		},
		"should successfully decode a Set": {
			typ: reflect.TypeOf(map[string]struct{}{}),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeArrayLen(2))
				assert.NoError(t, enc.EncodeInt(codeSet))
				assert.NoError(t, enc.EncodeArrayLen(2))
				assert.NoError(t, enc.EncodeString("a"))
				assert.NoError(t, enc.EncodeString("b"))
			},
			want:        map[string]struct{}{"a": {}, "b": {}},
			expectedErr: nil,
		},
		"should successfully decode bytes slice": {
			typ: reflect.TypeOf([]byte{}),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeArrayLen(2))
				assert.NoError(t, enc.EncodeInt(codeBytes))
				assert.NoError(t, enc.EncodeBytes([]byte{0x01, 0x02, 0x03}))
			},
			want:        []byte{0x01, 0x02, 0x03},
			expectedErr: nil,
		},
		"should successfully decode a Listing via slice": {
			typ: reflect.TypeOf([]int{}),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeArrayLen(2))
				assert.NoError(t, enc.EncodeInt(codeListing))
				assert.NoError(t, enc.EncodeArrayLen(2))
				assert.NoError(t, enc.EncodeInt(10))
				assert.NoError(t, enc.EncodeInt(20))
			},
			want:        []int{10, 20},
			expectedErr: nil,
		},
		"should successfully decode a Mapping via map": {
			typ: reflect.TypeOf(map[string]int{}),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeArrayLen(2))
				assert.NoError(t, enc.EncodeInt(codeMapping))
				assert.NoError(t, enc.EncodeMapLen(1))
				assert.NoError(t, enc.EncodeString("x"))
				assert.NoError(t, enc.EncodeInt(99))
			},
			want:        map[string]int{"x": 99},
			expectedErr: nil,
		},
		"should successfully decode Dynamic object": {
			typ: reflect.TypeOf(Object{}),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeArrayLen(4))
				assert.NoError(t, enc.EncodeInt(codeObject))
				assert.NoError(t, enc.EncodeString("Dynamic"))
				assert.NoError(t, enc.EncodeString("pkl:base"))
				// members array with 1 property, 1 entry, 1 element
				assert.NoError(t, enc.EncodeArrayLen(3))
				// property member
				assert.NoError(t, enc.EncodeArrayLen(3))
				assert.NoError(t, enc.EncodeInt(codeObjectMemberProperty))
				assert.NoError(t, enc.EncodeString("name"))
				assert.NoError(t, enc.EncodeString("Alice"))
				// entry member
				assert.NoError(t, enc.EncodeArrayLen(3))
				assert.NoError(t, enc.EncodeInt(codeObjectMemberEntry))
				assert.NoError(t, enc.EncodeString("key1"))
				assert.NoError(t, enc.EncodeString("val1"))
				// element member
				assert.NoError(t, enc.EncodeArrayLen(3))
				assert.NoError(t, enc.EncodeInt(codeObjectMemberElement))
				assert.NoError(t, enc.EncodeInt(0))
				assert.NoError(t, enc.EncodeString("elem0"))
			},
			want: Object{
				ModuleUri:  "pkl:base",
				Name:       "Dynamic",
				Properties: map[string]any{"name": "Alice"},
				Entries:    map[any]any{"key1": "val1"},
				Elements:   []any{"elem0"},
			},
			expectedErr: nil,
		},
		"should successfully decode interface with nil value": {
			typ: reflect.TypeOf((*interface{})(nil)).Elem(),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeNil())
			},
			want:        nil,
			expectedErr: nil,
		},
		"should successfully decode interface with bool value": {
			typ: reflect.TypeOf((*interface{})(nil)).Elem(),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeBool(true))
			},
			want:        true,
			expectedErr: nil,
		},
		"should successfully decode interface with int value": {
			typ: reflect.TypeOf((*interface{})(nil)).Elem(),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeInt(42))
			},
			want:        42,
			expectedErr: nil,
		},
		"should successfully decode interface with Pkl List": {
			typ: reflect.TypeOf((*interface{})(nil)).Elem(),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeArrayLen(2))
				assert.NoError(t, enc.EncodeInt(codeList))
				assert.NoError(t, enc.EncodeArrayLen(1))
				assert.NoError(t, enc.EncodeString("item"))
			},
			want:        []any{"item"},
			expectedErr: nil,
		},
		"should successfully decode interface with Pkl Set": {
			typ: reflect.TypeOf((*interface{})(nil)).Elem(),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeArrayLen(2))
				assert.NoError(t, enc.EncodeInt(codeSet))
				assert.NoError(t, enc.EncodeArrayLen(1))
				assert.NoError(t, enc.EncodeString("x"))
			},
			want:        map[any]any{"x": empty},
			expectedErr: nil,
		},
		"should successfully decode interface with Pkl Map": {
			typ: reflect.TypeOf((*interface{})(nil)).Elem(),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeArrayLen(2))
				assert.NoError(t, enc.EncodeInt(codeMap))
				assert.NoError(t, enc.EncodeMapLen(1))
				assert.NoError(t, enc.EncodeString("a"))
				assert.NoError(t, enc.EncodeString("b"))
			},
			want:        map[any]any{"a": "b"},
			expectedErr: nil,
		},
		"should return error for unsupported float32 type": {
			typ: reflect.TypeOf(float32(0)),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeFloat32(float32(0)))
			},
			expectedErr: &InternalError{
				err: fmt.Errorf("encountered unexpected Go kind while decoding: float32"),
			},
		},
		"should return error for invalid slice code": {
			typ: reflect.TypeOf([]string{}),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeArrayLen(2))
				assert.NoError(t, enc.EncodeInt(codeObject))
			},
			expectedErr: fmt.Errorf("invalid code for slices: %d. Expected %d or %d", codeObject, codeList, codeListing),
		},
		"should return error for invalid map code": {
			typ: reflect.TypeOf(map[string]string{}),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeArrayLen(2))
				assert.NoError(t, enc.EncodeInt(codeObject))
			},
			expectedErr: fmt.Errorf("invalid code for maps: %d", codeObject),
		},
		"should return error for unknown object code on interface": {
			typ: reflect.TypeOf((*interface{})(nil)).Elem(),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeArrayLen(2))
				assert.NoError(t, enc.EncodeInt(0x7F)) // a fake unknown code
			},
			expectedErr: &InternalError{
				err: fmt.Errorf("encountered unknown object code: %#02x", 0x7F),
			},
		},
		"should return error for struct with unknown code": {
			typ: reflect.TypeOf(dummyStruct{}),
			data: func(t *testing.T, enc *msgpack.Encoder) {
				assert.NoError(t, enc.EncodeArrayLen(2))
				assert.NoError(t, enc.EncodeInt(0x7F))
			},
			expectedErr: fmt.Errorf("code %#02x cannot be decoded into a struct", 0x7F),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			enc := msgpack.NewEncoder(&buf)
			if tc.data != nil {
				tc.data(t, enc)
			}
			d := newDecoder(buf.Bytes(), tc.schemas)

			got, err := d.Decode(tc.typ)
			if tc.expectedErr != nil {
				assert.Nil(t, got)
				assert.EqualError(t, err, tc.expectedErr.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, tc.want, got.Interface())
		})
	}
}
