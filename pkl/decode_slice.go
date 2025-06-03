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

func (d *decoder) decodeSlice(inType reflect.Type) (*reflect.Value, error) {
	length, code, err := d.decodeObjectPreamble()
	if err != nil {
		return nil, err
	}
	if length != 2 {
		return nil, fmt.Errorf("expected array length 2 but got %d", length)
	}
	if code != codeList && code != codeListing {
		return nil, fmt.Errorf("invalid code for slices: %d. Expected %d or %d", code, codeList, codeListing)
	}
	return d.decodeSliceImpl(inType)
}

func (d *decoder) decodeSliceImpl(inType reflect.Type) (*reflect.Value, error) {
	sliceLen, err := d.dec.DecodeArrayLen()
	if err != nil {
		return nil, err
	}
	elemType := inType.Elem()
	ret := reflect.MakeSlice(reflect.SliceOf(elemType), sliceLen, sliceLen)
	for i := 0; i < sliceLen; i++ {
		v := ret.Index(i)
		decoded, err := d.Decode(elemType)
		if err != nil {
			return nil, err
		}
		v.Set(*decoded)
	}
	return &ret, nil
}
