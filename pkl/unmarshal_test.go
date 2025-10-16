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

package pkl_test

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/apple/pkl-go/pkl"
	unknowntype "github.com/apple/pkl-go/pkl/test_fixtures/gen/unknown_type"

	any2 "github.com/apple/pkl-go/pkl/test_fixtures/gen/any"
	"github.com/apple/pkl-go/pkl/test_fixtures/gen/classes"
	"github.com/apple/pkl-go/pkl/test_fixtures/gen/collections"
	"github.com/apple/pkl-go/pkl/test_fixtures/gen/datasize"
	"github.com/apple/pkl-go/pkl/test_fixtures/gen/duration"
	"github.com/apple/pkl-go/pkl/test_fixtures/gen/dynamic"
	"github.com/apple/pkl-go/pkl/test_fixtures/gen/nullables"
	"github.com/apple/pkl-go/pkl/test_fixtures/gen/primitives"
	"github.com/apple/pkl-go/pkl/test_fixtures/gen/types"
	"github.com/apple/pkl-go/pkl/test_fixtures/gen/unions"
	"github.com/apple/pkl-go/pkl/test_fixtures/gen/unions/number"
	"github.com/apple/pkl-go/pkl/test_fixtures/gen/unions/othernumbers"
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"
)

//go:embed test_fixtures/msgpack/primitives.pkl.msgpack
var primitivesInput []byte

//go:embed test_fixtures/msgpack/collections.pkl.msgpack
var collectionsInput []byte

//go:embed test_fixtures/msgpack/duration.pkl.msgpack
var durationInput []byte

//go:embed test_fixtures/msgpack/datasize.pkl.msgpack
var datasizeInput []byte

//go:embed test_fixtures/msgpack/nullables.pkl.msgpack
var nullablesInput []byte

//go:embed test_fixtures/msgpack/dynamic.pkl.msgpack
var dynamicInput []byte

//go:embed test_fixtures/msgpack/classes.pkl.msgpack
var classesInput []byte

//go:embed test_fixtures/msgpack/unions.pkl.msgpack
var unionsInput []byte

// List(1, 2, 3)
//
//go:embed test_fixtures/msgpack/collections.res1.msgpack
var collectionsRes1 []byte

// new Listing { 2; 3; 4 }
//
//go:embed test_fixtures/msgpack/collections.res2.msgpack
var collectionsRes2 []byte

// Set("one", "two", "three")
//
//go:embed test_fixtures/msgpack/collections.res9.msgpack
var collectionsRes9 []byte

//go:embed test_fixtures/msgpack/any.pkl.msgpack
var anies []byte

//go:embed test_fixtures/msgpack/unknown_type.pkl.msgpack
var unknownType []byte

//go:embed test_fixtures/manual/arrays_too_long.pkl.msgpack
var arraysTooLong []byte

//go:embed test_fixtures/msgpack/types.pkl.msgpack
var typesInput []byte

//go:embed test_fixtures/manual/types_pre_0.30.pkl.msgpack
var typesPre030Input []byte

func TestUnmarshall_Primitives(t *testing.T) {
	var res primitives.Primitives
	expected := primitives.Primitives{
		Res0:  "bar",
		Res1:  1,
		Res2:  2,
		Res3:  3,
		Res4:  4,
		Res5:  5,
		Res6:  6,
		Res7:  7,
		Res8:  8,
		Res9:  5.3,
		Res10: true,
		Res11: nil,
		Res12: 33,
		Res13: 33.3333,
	}
	if assert.NoError(t, pkl.Unmarshal(primitivesInput, &res)) {
		assert.Equal(t, expected, res)
	}
}

func TestUnmarshall_Collections(t *testing.T) {
	var res collections.Collections
	expected := collections.Collections{
		Res1: []int{1, 2, 3},
		Res2: []int{2, 3, 4},
		Res3: [][]int{{1}, {2}, {3}},
		Res4: [][]int{{1}, {2}, {3}},
		Res5: map[int]bool{1: true, 2: false},
		Res6: map[int]map[int]bool{
			1: {
				1: true,
			},
			2: {
				2: true,
			},
			3: {
				3: true,
			},
		},
		Res7: map[int]bool{
			1: true,
			2: false,
		},
		Res8: map[int]map[int]bool{
			1: {
				1: true,
			},
			2: {
				2: false,
			},
		},
		Res9: map[string]struct{}{
			"one":   {},
			"two":   {},
			"three": {},
		},
		Res10: map[int8]struct{}{
			1: {},
			2: {},
			3: {},
		},
		Res11: pkl.Pair[int, float64]{
			First:  1,
			Second: 5.0,
		},
		Res12: pkl.Pair[any, any]{
			First:  "hello",
			Second: "goodbye",
		},
		Res13: pkl.Pair[int, *int]{
			First:  1,
			Second: &[]int{2}[0],
		},
		Res14: []byte{1, 2, 3, 4, 255},
	}
	if assert.NoError(t, pkl.Unmarshal(collectionsInput, &res)) {
		assert.Equal(t, expected, res)
	}
}

func TestUnmarshal_Duration(t *testing.T) {
	var res duration.Duration
	expected := duration.Duration{
		Res1: pkl.Duration{
			Value: 1,
			Unit:  pkl.Nanosecond,
		},
		Res2: pkl.Duration{
			Value: 2,
			Unit:  pkl.Microsecond,
		},
		Res3: pkl.Duration{
			Value: 3,
			Unit:  pkl.Millisecond,
		},
		Res4: pkl.Duration{
			Value: 4,
			Unit:  pkl.Second,
		},
		Res5: pkl.Duration{
			Value: 5,
			Unit:  pkl.Minute,
		},
		Res6: pkl.Duration{
			Value: 6,
			Unit:  pkl.Hour,
		},
		Res7: pkl.Duration{
			Value: 7,
			Unit:  pkl.Day,
		},
		Res8: pkl.Microsecond,
	}
	if assert.NoError(t, pkl.Unmarshal(durationInput, &res)) {
		assert.Equal(t, expected, res)
	}
}

func TestUnmarshal_DataSize(t *testing.T) {
	var res datasize.Datasize
	expected := datasize.Datasize{
		Res1:  pkl.DataSize{Value: 1, Unit: pkl.Bytes},
		Res2:  pkl.DataSize{Value: 2, Unit: pkl.Kilobytes},
		Res3:  pkl.DataSize{Value: 3, Unit: pkl.Megabytes},
		Res4:  pkl.DataSize{Value: 4, Unit: pkl.Gigabytes},
		Res5:  pkl.DataSize{Value: 5, Unit: pkl.Terabytes},
		Res6:  pkl.DataSize{Value: 6, Unit: pkl.Petabytes},
		Res7:  pkl.DataSize{Value: 7, Unit: pkl.Kibibytes},
		Res8:  pkl.DataSize{Value: 8, Unit: pkl.Mebibytes},
		Res9:  pkl.DataSize{Value: 9, Unit: pkl.Gibibytes},
		Res10: pkl.DataSize{Value: 10, Unit: pkl.Tebibytes},
		Res11: pkl.DataSize{Value: 11, Unit: pkl.Pebibytes},
		Res12: pkl.Megabytes,
	}
	if assert.NoError(t, pkl.Unmarshal(datasizeInput, &res)) {
		assert.Equal(t, expected, res)
	}
}

func TestUnmarshal_Nullables(t *testing.T) {
	var res nullables.Nullables
	expected := nullables.Nullables{
		Res0:  &[]string{"bar"}[0],
		Res1:  nil,
		Res2:  &[]int{1}[0],
		Res3:  nil,
		Res4:  &[]int8{2}[0],
		Res5:  nil,
		Res6:  &[]int16{3}[0],
		Res7:  nil,
		Res8:  &[]int32{4}[0],
		Res9:  nil,
		Res10: &[]uint{5}[0],
		Res11: nil,
		Res12: &[]uint8{6}[0],
		Res13: nil,
		Res14: &[]uint16{7}[0],
		Res15: nil,
		Res16: &[]uint32{8}[0],
		Res17: nil,
		Res18: &[]float64{5.3}[0],
		Res19: nil,
		Res20: &[]bool{true}[0],
		Res21: nil,
		Res22: &map[string]string{"foo": "bar"},
		Res23: nil,
		// can't test this due to https://github.com/stretchr/testify/issues/1143
		//Res24: &map[*string]*string{
		//	&[]string{"foo"}[0]:  &[]string{"bar"}[0],
		//	nil:                  nil,
		//	&[]string{"foo2"}[0]: nil,
		//},
		Res25: nil,
		Res26: &[]*int{
			&[]int{1}[0],
			&[]int{2}[0],
			nil,
			&[]int{4}[0],
			&[]int{5}[0],
		},
		Res27: nil,
		Res28: &nullables.MyClass{Prop: nil},
		Res29: &nullables.MyClass{Prop: &[]string{"foo"}[0]},
		Res30: nil,
	}
	if assert.NoError(t, pkl.Unmarshal(nullablesInput, &res)) {
		assert.Equal(t, expected, res)
	}
}

func TestUnmarshal_Dynamic(t *testing.T) {
	var res dynamic.Dynamic
	expected := dynamic.Dynamic{
		Res1: pkl.Object{
			ModuleUri: "pkl:base",
			Name:      "Dynamic",
			Properties: map[string]any{
				"res2": pkl.Object{
					ModuleUri:  "pkl:base",
					Name:       "Dynamic",
					Properties: map[string]any{"res3": 5},
					Entries:    map[any]any{},
					Elements:   []any{},
				},
				"res5": dynamic.MyClass{MyValue: 8},
			},
			Entries: map[any]any{
				"res4": 6,
				5:      9,
			},
			Elements: []any{
				dynamic.MyClass{
					MyValue: 7,
				},
			},
		},
	}
	if assert.NoError(t, pkl.Unmarshal(dynamicInput, &res)) {
		assert.Equal(t, expected, res)
	}
}

func TestUnmarshal_Classes(t *testing.T) {
	var greyhound = classes.GreyhoundImpl{
		DogImpl: classes.DogImpl{
			Name:  "Uni",
			Barks: false,
			Breed: "Greyhound",
		},
		CanRoach: true,
	}
	var animal classes.Animal = greyhound
	expected := classes.Classes{
		Animals: []classes.Animal{
			greyhound,
			classes.CatImpl{
				Name:  "Millie",
				Meows: true,
			},
		},
		NullableAnimals: []*classes.Animal{
			nil,
			&animal,
		},
		MyAnimal: classes.GreyhoundImpl{
			DogImpl: classes.DogImpl{
				Name:  "Uni",
				Barks: false,
				Breed: "Greyhound",
			},
			CanRoach: true,
		},
		House: classes.House{
			Area:      2000,
			Bedrooms:  3,
			Bathrooms: 2,
		},
	}
	var res classes.Classes
	if assert.NoError(t, pkl.Unmarshal(classesInput, &res)) {
		assert.Equal(t, expected, res)
	}
}

func TestUnmarshal_Unions(t *testing.T) {
	expected := unions.Unions{
		Res1: number.Two,
		Res2: othernumbers.N三,
	}
	var res unions.Unions
	if assert.NoError(t, pkl.Unmarshal(unionsInput, &res)) {
		assert.Equal(t, expected, res)
	}
}

func TestUnmarshal_Strings(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	enc := msgpack.NewEncoder(buf)
	if assert.NoError(t, enc.Encode("hello")) {
		var res string
		if assert.NoError(t, pkl.Unmarshal(buf.Bytes(), &res)) {
			assert.Equal(t, "hello", res)
		}
	}
}

func TestUnmarshal_Int(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	enc := msgpack.NewEncoder(buf)
	if assert.NoError(t, enc.Encode(5)) {
		var res int
		if assert.NoError(t, pkl.Unmarshal(buf.Bytes(), &res)) {
			assert.Equal(t, 5, res)
		}
	}
}

func TestUnmarshal_Slice(t *testing.T) {
	var res []int
	if assert.NoError(t, pkl.Unmarshal(collectionsRes1, &res)) {
		assert.Equal(t, []int{1, 2, 3}, res)
	}
	if assert.NoError(t, pkl.Unmarshal(collectionsRes2, &res)) {
		assert.Equal(t, []int{2, 3, 4}, res)
	}
	var badRes []string
	assert.Error(t, pkl.Unmarshal(collectionsRes1, &badRes))
}

func TestUnmarshal_Set(t *testing.T) {
	var res map[string]struct{}
	if assert.NoError(t, pkl.Unmarshal(collectionsRes9, &res)) {
		assert.Equal(t, map[string]struct{}{
			"one":   {},
			"two":   {},
			"three": {},
		}, res)
	}
}

func TestUnmarshal_Set_any(t *testing.T) {
	var res any
	emptyStruct := struct{}{}
	if assert.NoError(t, pkl.Unmarshal(collectionsRes9, &res)) {
		assert.Equal(t, map[any]any{
			"one":   emptyStruct,
			"two":   emptyStruct,
			"three": emptyStruct,
		}, res)
	}
}

func TestUnmarshal_AnyType(t *testing.T) {
	var res any2.Any
	err := pkl.Unmarshal(anies, &res)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, any2.Any{
		Res1: []any{
			any2.Person{
				Name: "Barney",
			},
		},
		Res2: any2.Person{
			Name: "Bobby",
		},
		Res3: map[any]any{
			"Wilma": any2.Person{
				Name: "Wilma",
			},
		},
	}, res)
}

func TestUnmarshal_UnknownType(t *testing.T) {
	var res unknowntype.UnknownType
	err := pkl.Unmarshal(unknownType, &res)
	assert.Error(t, err)
	assert.Equal(t, "cannot decode Pkl value of type `PcfRenderer` into Go type `interface {}`. Define a custom mapping for this using `pkl.RegisterMapping`", err.Error())
}

func TestUnmarshal_ArraysTooLong(t *testing.T) {
	var res pkl.Object
	assert.NoError(t, pkl.Unmarshal(arraysTooLong, &res))
}

func TestUnmarshal_Types(t *testing.T) {
	var res types.Types
	assert.NoError(t, pkl.Unmarshal(typesInput, &res))

	assert.Equal(t, types.Types{
		StringClass:     pkl.Class{ModuleUri: "pkl:base", Name: "String"},
		BaseModuleClass: pkl.Class{ModuleUri: "pkl:base", Name: "ModuleClass"},
		Uint8TypeAlias:  pkl.TypeAlias{ModuleUri: "pkl:base", Name: "UInt8"},
		FooClass:        pkl.Class{ModuleUri: "pklgo:/pkl/test_fixtures/types.pkl", Name: "types#Foo"},
		BarTypeAlias:    pkl.TypeAlias{ModuleUri: "pklgo:/pkl/test_fixtures/types.pkl", Name: "types#Bar"},
	}, res)
}

func TestUnmarshal_Types_Pre_030(t *testing.T) {
	var res types.Types
	assert.NoError(t, pkl.Unmarshal(typesPre030Input, &res))

	assert.Equal(t, types.Types{}, res)
}
