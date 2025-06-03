// ===----------------------------------------------------------------------===//
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
// ===----------------------------------------------------------------------===//

package pkl

import (
	"fmt"
	"reflect"
)

func (d *decoder) decodeMap(inType reflect.Type) (*reflect.Value, error) {
	_, code, err := d.decodeObjectPreamble()
	if err != nil {
		return nil, err
	}
	if code == codeSet {
		return d.decodeSet(inType)
	}
	if code != codeMap && code != codeMapping {
		return nil, fmt.Errorf("invalid code for maps: %d", code)
	}
	return d.decodeMapImpl(inType)
}

func (d *decoder) decodeMapImpl(inType reflect.Type) (*reflect.Value, error) {
	mapLen, err := d.dec.DecodeMapLen()
	if err != nil {
		return nil, err
	}
	ret := reflect.MakeMapWithSize(inType, mapLen)
	keyType := inType.Key()
	valueType := inType.Elem()
	for i := 0; i < mapLen; i++ {
		key, err := d.Decode(keyType)
		if err != nil {
			return nil, err
		}
		value, err := d.Decode(valueType)
		if err != nil {
			return nil, err
		}
		ret.SetMapIndex(*key, *value)
	}
	return &ret, nil
}

var emptyMirror = reflect.ValueOf(empty)

// decodeSet decodes into `map[T]struct{}`
func (d *decoder) decodeSet(inType reflect.Type) (*reflect.Value, error) {
	length, err := d.dec.DecodeArrayLen()
	if err != nil {
		return nil, err
	}
	ret := reflect.MakeMapWithSize(inType, length)
	keyType := inType.Key()
	for i := 0; i < length; i++ {
		elem, err := d.Decode(keyType)
		if err != nil {
			return nil, err
		}
		ret.SetMapIndex(*elem, emptyMirror)
	}
	return &ret, nil
}
