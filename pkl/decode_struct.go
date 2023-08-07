// ===----------------------------------------------------------------------===//
// Copyright © 2024 Apple Inc. and the Pkl project authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ===----------------------------------------------------------------------===//
package pkl

import (
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/vmihailenco/msgpack/v5/msgpcode"
)

const StructTag = "pkl"

type structFieldOpts struct {
	propertyName string
}

type structField struct {
	*reflect.StructField
	structFieldOpts
}

var objectType = reflect.TypeOf(Object{})

var sliceOfEmptyInterface []interface{}
var emptyInterfaceType = reflect.TypeOf(sliceOfEmptyInterface).Elem()

// decodeStruct decodes into an object represented by typ.
// If outValue is not nil, writes fields onto outValue.
func (d *decoder) decodeStruct(typ reflect.Type) (*reflect.Value, error) {
	_, code, err := d.decodeObjectPreamble()
	if err != nil {
		return nil, err
	}
	switch code {
	case codeObject:
		return d.decodeObject(typ)
	case codeDataSize:
		return d.decodeDataSize()
	case codeDuration:
		return d.decodeDuration()
	case codePair:
		return d.decodePair(typ)
	case codeIntSeq:
		return d.decodeIntSeq()
	case codeRegex:
		return d.decodeRegex()
	case codeClass:
		ret := reflect.ValueOf(&Class{})
		return &ret, nil
	case codeTypeAlias:
		ret := reflect.ValueOf(&TypeAlias{})
		return &ret, nil
	default:
		return nil, fmt.Errorf("code %x cannot be decoded into a struct", code)
	}
}

func (d *decoder) decodeObject(typ reflect.Type) (*reflect.Value, error) {
	name, err := d.dec.DecodeString()
	if err != nil {
		return nil, err
	}
	moduleUri, err := d.dec.DecodeString()
	if err != nil {
		return nil, err
	}
	if moduleUri == "pkl:base" && name == "Dynamic" || typ.AssignableTo(objectType) {
		return d.decodeObjectGeneric(moduleUri, name)
	}
	return d.decodeTyped(name, typ)
}

// decodeObjectGeneric decodes into Object.
func (d *decoder) decodeObjectGeneric(moduleUri, name string) (*reflect.Value, error) {
	obj := Object{
		ModuleUri:  moduleUri,
		Name:       name,
		Properties: make(map[string]any),
		Entries:    make(map[any]any),
		Elements:   []any{},
	}
	length, err := d.dec.DecodeArrayLen()
	if err != nil {
		return nil, err
	}
	for i := 0; i < length; i++ {
		_, err := d.dec.DecodeArrayLen()
		if err != nil {
			return nil, err
		}
		code, err := d.dec.DecodeInt()
		if err != nil {
			return nil, err
		}
		switch code {
		case codeObjectMemberProperty:
			name, err := d.dec.DecodeString()
			if err != nil {
				return nil, err
			}
			value, err := d.decodeInterface(emptyInterfaceType)
			if err != nil {
				return nil, err
			}
			obj.Properties[name] = value.Interface()
		case codeObjectMemberEntry:
			key, err := d.decodeInterface(emptyInterfaceType)
			if err != nil {
				return nil, err
			}
			value, err := d.decodeInterface(emptyInterfaceType)
			if err != nil {
				return nil, err
			}
			obj.Entries[key.Interface()] = value.Interface()
		case codeObjectMemberElement:
			// index
			_, err := d.dec.DecodeInt()
			if err != nil {
				return nil, err
			}
			value, err := d.decodeInterface(emptyInterfaceType)
			if err != nil {
				return nil, err
			}
			obj.Elements = append(obj.Elements, value.Interface())
		}
	}
	ret := reflect.ValueOf(obj)
	return &ret, nil
}

func (d *decoder) decodeTyped(name string, typ reflect.Type) (*reflect.Value, error) {
	if t, exists := d.schemas[name]; exists {
		// if we have a known schema by name, use that type instead of the input typ.
		// this is important if the Pkl value is a subtype of the input type, e.g.
		// in polymorphic cases.
		typ = t
	} else if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("cannot decode Pkl value of type `%s` into Go type `%s`. Define a custom mapping for this using `pkl.RegisterMapping`", name, typ)
	}
	out, err := d.getOutputValue(typ)
	if err != nil {
		return nil, err
	}
	propertiesLen, err := d.dec.DecodeArrayLen()
	if err != nil {
		return nil, err
	}
	fields := getStructFields(typ)
	for i := 0; i < propertiesLen; i++ {
		if err = d.decodeStructField(fields, out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (d *decoder) decodeDuration() (*reflect.Value, error) {
	value, err := d.dec.DecodeFloat64()
	if err != nil {
		return nil, err
	}
	unitStr, err := d.dec.DecodeString()
	if err != nil {
		return nil, err
	}
	unit, err := ToDurationUnit(unitStr)
	if err != nil {
		return nil, err
	}
	ret := reflect.ValueOf(Duration{
		Unit:  unit,
		Value: value,
	})
	return &ret, nil
}

func (d *decoder) decodeDataSize() (*reflect.Value, error) {
	value, err := d.dec.DecodeFloat64()
	if err != nil {
		return nil, err
	}
	unitStr, err := d.dec.DecodeString()
	if err != nil {
		return nil, err
	}
	unit, err := ToDataSizeUnit(unitStr)
	if err != nil {
		return nil, err
	}
	ret := reflect.ValueOf(DataSize{
		Unit:  unit,
		Value: value,
	})
	return &ret, nil
}

func (d *decoder) decodePair(typ reflect.Type) (*reflect.Value, error) {
	firstField, exists := typ.FieldByName("First")
	if !exists {
		return nil, &InternalError{
			err: errors.New("unable to find field `First` on pkl.Pair"),
		}
	}
	first, err := d.Decode(firstField.Type)
	if err != nil {
		return nil, err
	}
	secondField, exists := typ.FieldByName("Second")
	if !exists {
		if !exists {
			return nil, &InternalError{
				err: errors.New("unable to find field `Second` on pkl.Pair"),
			}
		}
	}
	second, err := d.Decode(secondField.Type)
	if err != nil {
		return nil, err
	}
	ret := reflect.New(typ)
	elem := ret.Elem()
	elem.FieldByName("First").Set(*first)
	elem.FieldByName("Second").Set(*second)
	return &elem, nil
}

func (d *decoder) decodeStructField(fields map[string]structField, out *reflect.Value) error {
	if _, err := d.dec.DecodeArrayLen(); err != nil {
		return err
	}
	memberCode, err := d.dec.DecodeInt()
	if err != nil {
		return err
	}
	if memberCode != codeObjectMemberProperty {
		return fmt.Errorf("expected code %d but found %d", codeObjectMemberProperty, memberCode)
	}
	propertyName, err := d.dec.DecodeString()
	if err != nil {
		return err
	}
	sf, exists := fields[propertyName]
	if !exists {
		log.Default().Printf("warn: Cannot find field on Go struct `%s` matching Pkl property `%s`. Ensure the Go structs are up to date with Pkl classes either through codegen or manually adding `pkl` tags.", out.Type().String(), propertyName)
		return d.dec.Skip()
	}
	code, err := d.dec.PeekCode()
	if err != nil {
		return err
	}
	// If value is nil, the struct field's value is already nil because it is the zero value.
	if code == msgpcode.Nil {
		return d.dec.Skip()
	}
	decodedValue, err := d.Decode(sf.Type)
	if err != nil {
		return err
	}
	out.FieldByName(sf.Name).Set(*decodedValue)
	return nil
}

func (d *decoder) decodeClass() (*reflect.Value, error) {
	ret := reflect.ValueOf(&Class{})
	return &ret, nil
}

func (d *decoder) decodeTypeAlias() (*reflect.Value, error) {
	ret := reflect.ValueOf(&TypeAlias{})
	return &ret, nil
}

func parseStructOpts(field *reflect.StructField) structFieldOpts {
	ret := structFieldOpts{propertyName: field.Name}
	tagValue, exists := field.Tag.Lookup(StructTag)
	if !exists {
		return ret
	}
	ret.propertyName = tagValue
	return ret
}

func getStructFields(typ reflect.Type) map[string]structField {
	numFields := typ.NumField()
	ret := make(map[string]structField)
	for i := 0; i < numFields; i++ {
		field := typ.Field(i)
		// embedded
		if field.Anonymous {
			for k, v := range getStructFields(field.Type.Elem()) {
				ret[k] = v
			}
		} else {
			opts := parseStructOpts(&field)
			if opts.propertyName == "-" {
				continue
			}
			ret[opts.propertyName] = structField{StructField: &field, structFieldOpts: opts}
		}
	}
	return ret
}

// Returns the output value to write into.
func (d *decoder) getOutputValue(typ reflect.Type) (*reflect.Value, error) {
	ret := reflect.New(typ).Elem()
	// initialize all embedded structs.
	numFields := typ.NumField()
	for i := 0; i < numFields; i++ {
		field := typ.Field(i)
		if field.Anonymous {
			fieldValue := reflect.New(field.Type.Elem())
			// Assertion: all embedded fields are pointers to structs.
			structValue, err := d.getOutputValue(field.Type.Elem())
			if err != nil {
				return nil, err
			}
			fieldValue.Elem().Set(*structValue)
			ret.FieldByName(field.Name).Set(fieldValue)
		}
	}
	return &ret, nil
}

func (d *decoder) decodeIntSeq() (*reflect.Value, error) {
	start, err := d.dec.DecodeInt()
	if err != nil {
		return nil, err
	}
	end, err := d.dec.DecodeInt()
	if err != nil {
		return nil, err
	}
	step, err := d.dec.DecodeInt()
	if err != nil {
		return nil, err
	}
	intseq := IntSeq{Start: start, End: end, Step: step}
	ret := reflect.ValueOf(intseq)
	return &ret, nil
}

func (d *decoder) decodeRegex() (*reflect.Value, error) {
	pattern, err := d.dec.DecodeString()
	if err != nil {
		return nil, err
	}
	regexp := Regex{Pattern: pattern}
	ret := reflect.ValueOf(regexp)
	return &ret, nil
}
