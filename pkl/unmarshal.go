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
	"errors"
	"fmt"
	"reflect"
)

// Unmarshal parses Pkl-encoded data and stores the result into
// the value pointed by v.
//
// This is a low-level API. Most users should be using Evaluator.Evaluate instead.
//
// The following struct tags are supported:
//
//	pkl:"Field"     Overrides the field's name to map to.
//
//goland:noinspection GoUnusedExportedFunction
func Unmarshal(data []byte, v any) error {
	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Ptr {
		return fmt.Errorf("cannot unmarshal non-pointer. Got kind: %v", value.Kind())
	}
	if value.IsNil() {
		return errors.New("cannot unmarshal into nil")
	}
	res, err := newDecoder(data, schemas).Decode(value.Elem().Type())
	if err != nil {
		return err
	}
	value.Elem().Set(*res)
	return nil
}
