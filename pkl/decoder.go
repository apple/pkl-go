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
	"bytes"
	"encoding"
	"fmt"
	"reflect"
	"time"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/vmihailenco/msgpack/v5/msgpcode"
)

const (
	codeObject               = 0x1
	codeMap                  = 0x2
	codeMapping              = 0x3
	codeList                 = 0x4
	codeListing              = 0x5
	codeSet                  = 0x6
	codeDuration             = 0x7
	codeDataSize             = 0x8
	codePair                 = 0x9
	codeIntSeq               = 0xA
	codeRegex                = 0xB
	codeClass                = 0xC
	codeTypeAlias            = 0xD
	codeObjectMemberProperty = 0x10
	codeObjectMemberEntry    = 0x11
	codeObjectMemberElement  = 0x12
)

type decoder struct {
	dec     *msgpack.Decoder
	schemas map[string]reflect.Type
}

func newDecoder(b []byte, schemas map[string]reflect.Type) *decoder {
	msgpackDecoder := msgpack.NewDecoder(bytes.NewReader(b))
	return &decoder{
		dec:     msgpackDecoder,
		schemas: schemas,
	}
}

var durationType = reflect.TypeOf(time.Duration(0))

// Decode decodes the next value according to the expected type.
func (d *decoder) Decode(typ reflect.Type) (res *reflect.Value, err error) {
	res, isUnmarshaled, err := d.maybeUnmarshal(typ)
	if isUnmarshaled {
		return res, err
	}
	switch typ.Kind() {
	case reflect.Ptr:
		return d.decodePointer(typ)
	case reflect.Struct:
		return d.decodeStruct(typ)
	case reflect.Bool:
		return d.decodeBool()
	case reflect.String:
		return d.decodeString()
	case reflect.Int:
		return d.decodeInt()
	case reflect.Int8:
		return d.decodeInt8()
	case reflect.Int16:
		return d.decodeInt16()
	case reflect.Int32:
		return d.decodeInt32()
	case reflect.Int64:
		switch typ {
		case durationType:
			return d.decodeDuration()
		default:
			return d.decodeInt64()
		}
	case reflect.Uint:
		return d.decodeUint()
	case reflect.Uint8:
		return d.decodeUint8()
	case reflect.Uint16:
		return d.decodeUint16()
	case reflect.Uint32:
		return d.decodeUint32()
	case reflect.Uint64:
		return d.decodeUint64()
	case reflect.Float64:
		return d.decodeFloat64()
	case reflect.Slice:
		return d.decodeSlice(typ)
	case reflect.Map:
		return d.decodeMap(typ)
	case reflect.Interface:
		return d.decodeInterface(typ)
	default:
		return nil, &InternalError{
			err: fmt.Errorf("encountered unexpected Go kind while decoding: %s", typ.Kind()),
		}
	}
}

func (d *decoder) decodePointer(inType reflect.Type) (*reflect.Value, error) {
	code, err := d.dec.PeekCode()
	if err != nil {
		return nil, err
	}
	if code == msgpcode.Nil {
		if err = d.dec.Skip(); err != nil {
			return nil, err
		}
		ret := reflect.Zero(inType)
		return &ret, nil
	}
	val, err := d.Decode(inType.Elem())
	if err != nil {
		return nil, err
	}
	// if the decoded is already a pointer, we can just return it
	if val.Kind() == reflect.Ptr {
		return val, nil
	}
	ret := reflect.New(inType.Elem())
	ret.Elem().Set(*val)
	return &ret, nil
}

// maybeUnmarshal determines if typ implements encoding.BinaryUnmarshaler, and if it does,
// performs the unmarshaling.
func (d *decoder) maybeUnmarshal(typ reflect.Type) (*reflect.Value, bool, error) {
	ptr := reflect.New(typ)
	unmarshaler, isUnmarshaler := ptr.Interface().(encoding.BinaryUnmarshaler)
	if !isUnmarshaler {
		return nil, false, nil
	}
	b, err := d.dec.DecodeBytes()
	if err != nil {
		return nil, true, err
	}
	if err = unmarshaler.UnmarshalBinary(b); err != nil {
		return nil, true, err
	}
	ret := ptr.Elem()
	return &ret, true, nil
}

// decodeInterface decodes for an input type that is an interface.
func (d *decoder) decodeInterface(typ reflect.Type) (*reflect.Value, error) {
	code, err := d.dec.PeekCode()
	if err != nil {
		return nil, err
	}
	var ret reflect.Value
	switch {
	case msgpcode.IsFixedMap(code) || code == msgpcode.Map16 || code == msgpcode.Map32:
		return d.decodeMapImpl(typ)
	case msgpcode.IsFixedArray(code):
		return d.decodePklObject(typ)
	case msgpcode.IsString(code):
		return d.decodeString()
	case code == msgpcode.Nil:
		if err = d.dec.Skip(); err != nil {
			return nil, err
		}
		// the zero value of interface{} is `nil`.
		ret = reflect.Zero(typ)
	default:
		// the rest are primitive types.
		// All integers become `int`.
		value, err := d.dec.DecodeInterfaceLoose()
		if err != nil {
			return nil, err
		}
		switch value := value.(type) {
		case int64:
			ret = reflect.ValueOf(int(value))
		case uint64:
			ret = reflect.ValueOf(int(value))
		default:
			ret = reflect.ValueOf(value)
		}
	}
	return &ret, nil
}

func (d *decoder) decodePklObject(typ reflect.Type) (*reflect.Value, error) {
	_, err := d.dec.DecodeArrayLen()
	if err != nil {
		return nil, err
	}
	code, err := d.dec.DecodeInt()
	if err != nil {
		return nil, err
	}
	switch code {
	case codeObject:
		return d.decodeObject(typ)
	case codeMap:
		fallthrough
	case codeMapping:
		return d.decodeMapImpl(reflect.TypeOf(map[any]any{}))
	case codeList:
		fallthrough
	case codeListing:
		return d.decodeSliceImpl(reflect.TypeOf([]any{}))
	case codeSet:
		return d.decodeSet(reflect.TypeOf(map[any]any{}))
	case codeDataSize:
		return d.decodeDataSize()
	case codeDuration:
		return d.decodeDuration()
	case codeIntSeq:
		return d.decodeIntSeq()
	case codeRegex:
		return d.decodeRegex()
	case codeClass:
		return d.decodeClass()
	case codeTypeAlias:
		return d.decodeTypeAlias()
	default:
		return nil, &InternalError{
			err: fmt.Errorf("encountered unknown object code: %d", code),
		}
	}
}

// decodeObjectPreamble decodes the preamble for Pkl objects.
//
// All Pkl objects are packed as array types, where the first slot is a code
// that describes the type of object.
func (d *decoder) decodeObjectPreamble() (int, int, error) {
	arrLen, err := d.dec.DecodeArrayLen()
	if err != nil {
		return 0, 0, err
	}
	code, err := d.dec.DecodeInt()
	if err != nil {
		return 0, 0, err
	}
	return arrLen, code, err
}
