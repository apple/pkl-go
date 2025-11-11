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

//goland:noinspection GoUnusedConst
const (
	codeObject               = 0x01
	codeMap                  = 0x02
	codeMapping              = 0x03
	codeList                 = 0x04
	codeListing              = 0x05
	codeSet                  = 0x06
	codeDuration             = 0x07
	codeDataSize             = 0x08
	codePair                 = 0x09
	codeIntSeq               = 0x0A
	codeRegex                = 0x0B
	codeClass                = 0x0C
	codeTypeAlias            = 0x0D
	codeFunction             = 0x0E
	codeBytes                = 0x0F
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
	res, isUnmarshalled, err := d.maybeUnmarshal(typ)
	if isUnmarshalled {
		return res, err
	}
	switch typ.Kind() {
	case reflect.Ptr:
		return d.decodePointer(typ)
	case reflect.Struct:
		return d.decodePklObject(typ, true)
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
	switch {
	case val.Type().AssignableTo(ret.Elem().Type()):
		ret.Elem().Set(*val)
	case val.Type().ConvertibleTo(ret.Elem().Type()):
		ret.Elem().Set(val.Convert(ret.Elem().Type()))
	default:
		return nil, fmt.Errorf("unable to assign or convert value of type `%s` to pointer of type `%s`", val.Type(), ret.Type())
	}
	return &ret, nil
}

// maybeUnmarshal determines if typ implements encoding.BinaryUnmarshaler, and if it does,
// performs the unmarshalling.
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
		return d.decodePklObject(typ, false)
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

func (d *decoder) decodePklObject(typ reflect.Type, requireStruct bool) (res *reflect.Value, err error) {
	length, code, err := d.decodeObjectPreamble()
	if err != nil {
		return nil, err
	}
	switch {
	case code == codeObject:
		res, err = d.decodeObject(typ)
	case !requireStruct && (code == codeMap || code == codeMapping):
		res, err = d.decodeMapImpl(reflect.TypeOf(map[any]any{}))
	case !requireStruct && (code == codeList || code == codeListing):
		res, err = d.decodeSliceImpl(reflect.TypeOf([]any{}))
	case !requireStruct && code == codeSet:
		res, err = d.decodeSet(reflect.TypeOf(map[any]any{}))
	case code == codeDataSize:
		res, err = d.decodeDataSize()
	case code == codeDuration:
		res, err = d.decodeDuration()
	case code == codePair:
		if typ == emptyInterfaceType {
			res, err = d.decodePair(reflect.TypeOf(Pair[any, any]{}))
		} else {
			res, err = d.decodePair(typ)
		}
	case code == codeIntSeq:
		res, err = d.decodeIntSeq()
	case code == codeRegex:
		res, err = d.decodeRegex()
	case code == codeClass:
		res, err = d.decodeClass(length)
	case code == codeTypeAlias:
		res, err = d.decodeTypeAlias(length)
	default:
		if requireStruct {
			return nil, fmt.Errorf("code %#02x cannot be decoded into a struct", code)
		}
		return nil, &InternalError{
			err: fmt.Errorf("encountered unknown object code: %#02x", code),
		}
	}

	if err != nil {
		return nil, err
	}
	return res, d.skip(length - getDecodedLength(code, length) - 1) // -1 is from the code field
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

// getDecodedLength returns the number of array fields a specific type code is expected to yield
func getDecodedLength(code, length int) int {
	switch code {
	case codeObject:
		return 3 // name, moduleUri, member array
	case codeDataSize, codeDuration:
		return 2 // value, unitStr
	case codePair:
		return 2 // first, second
	case codeIntSeq:
		return 3 // start, end, step
	case codeClass, codeTypeAlias:
		if length > 1 {
			// pkl 0.30+ includes the qualified name and module uri
			return 2
		}
		// before pkl 0.30 only the type code is present
		return 0
	default:
		return 1
	}
}

// skip provides a utility for ensuring forward-compatibility of fixed-size array types.
// Any time something is decoded from an array of some expected fixed size this should be called
// with `<array length> - <actual decoded value count>` as the argument.
func (d *decoder) skip(length int) error {
	if length < 0 {
		panic("skip length < 0")
	}
	for range length {
		if err := d.dec.Skip(); err != nil {
			return err
		}
	}
	return nil
}
