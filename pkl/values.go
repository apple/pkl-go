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
	"encoding"
	"fmt"
	"strconv"
	"time"
)

// Object is the Go representation of `pkl.base#Object`.
// It is a container for properties, entries, and elements.
type Object struct {
	// ModuleUri is the URI of the module that holds the definition of this object's class.
	ModuleUri string

	// Name is the qualified name of Pkl object's class.
	//
	// Example:
	//
	// 		"pkl.base#Dynamic"
	Name string

	// Properties is the set of name-value pairs in an Object.
	Properties map[string]any

	// Entries is the set of key-value pairs in an Object.
	Entries map[any]any

	// Elements is the set of items in an Object
	Elements []any
}

// Pair is the Go representation of `pkl.base#Pair`.
//
// It is an ordered pair of elements.
type Pair[A any, B any] struct {
	// First is the first element of the pair.
	First A

	// Second is the second element of the pair.
	Second B
}

// Regex is the Go representation of `pkl.base#Regex`.
//
// Regulard experssions in Pkl are
type Regex struct {
	// Pattern is the regex pattern expression in string form.
	Pattern string
}

// Class is the Go representation of `pkl.base#Class`.
//
// This value is purposefully opaque, and only exists for compatibilty.
type Class struct{}

// TypeAlias is the Go representation of `pkl.base#TypeAlias`.
//
// This value is purposefully opaque, and only exists for compatibilty.
type TypeAlias struct{}

// IntSeq is the Go representation of `pkl.base#IntSeq`.
//
// This value exists for compatibility. IntSeq should preferrably be used as a way to describe
// logic within a Pkl program, and not passed as data between Pkl and Go.
type IntSeq struct {
	// Start is the start of this seqeunce.
	Start int

	// End is the end of this seqeunce.
	End int

	// Step is the common difference of successive members of this sequence.
	Step int
}

type Duration struct {
	Value float64

	Unit DurationUnit
}

// GoDuration returns the duration as a time.Duration.
func (d *Duration) GoDuration() time.Duration {
	return time.Duration(d.Value * float64(d.Unit))
}

// DurationUnit represents unit of a Duration.
type DurationUnit int64

const (
	Nanosecond  DurationUnit = 1
	Microsecond              = Nanosecond * 1000
	Millisecond              = Microsecond * 1000
	Second                   = Millisecond * 1000
	Minute                   = Second * 60
	Hour                     = Minute * 60
	Day                      = Hour * 24
)

var _ encoding.BinaryUnmarshaler = new(DurationUnit)

// String returns the string representation of this DataSizeUnit.
//
//goland:noinspection GoMixedReceiverTypes
func (d DurationUnit) String() string {
	switch d {
	case Nanosecond:
		return "ns"
	case Microsecond:
		return "us"
	case Millisecond:
		return "ms"
	case Second:
		return "s"
	case Minute:
		return "min"
	case Hour:
		return "h"
	case Day:
		return "d"
	default:
		return "<invalid>"
	}
}

//goland:noinspection GoMixedReceiverTypes
func (d *DurationUnit) UnmarshalBinary(data []byte) error {
	unit, err := ToDurationUnit(string(data))
	if err != nil {
		return err
	}
	*d = unit
	return nil
}

// ToDurationUnit converts to a DurationUnit from its string representation.
func ToDurationUnit(str string) (DurationUnit, error) {
	switch str {
	case "ns":
		return Nanosecond, nil
	case "us":
		return Microsecond, nil
	case "ms":
		return Millisecond, nil
	case "s":
		return Second, nil
	case "min":
		return Minute, nil
	case "h":
		return Hour, nil
	case "d":
		return Day, nil
	default:
		return 0, fmt.Errorf("unrecognized Duration unit: `%s`", str)
	}
}

// DataSize is the Go representation of `pkl.base#DataSize`.
//
// It represents a quantity of binary data, represented by Value (e.g. 30.5) and Unit
// (e.g. Megabytes).
type DataSize struct {
	// Value is the value of this data size.
	Value float64

	// Unit is the unit of this data size.
	Unit DataSizeUnit
}

// String implementers the fmt.Stringer interface for DataSize.
func (d *DataSize) String() string {
	value := strconv.FormatFloat(d.Value, 'f', -1, 64)
	return fmt.Sprintf("%s.%s", value, d.Unit.String())
}

// ToUnit converts this DataSize to the specified unit.
func (d *DataSize) ToUnit(unit DataSizeUnit) DataSize {
	return DataSize{
		Unit:  unit,
		Value: d.Value / float64(unit),
	}
}

// DataSizeUnit represents unit of a DataSize.
type DataSizeUnit int64

var _ encoding.BinaryUnmarshaler = new(DataSizeUnit)

const (
	Bytes     DataSizeUnit = 1
	Kilobytes              = 1000
	Kibibytes              = 1024
	Megabytes              = Kilobytes * 1000
	Mebibytes              = Kibibytes * 1024
	Gigabytes              = Megabytes * 1000
	Gibibytes              = Mebibytes * 1024
	Terabytes              = Gigabytes * 1000
	Tebibytes              = Gibibytes * 1024
	Petabytes              = Terabytes * 1000
	Pebibytes              = Tebibytes * 1024
)

// String returns the string representation of this DataSizeUnit.
//
//goland:noinspection GoMixedReceiverTypes
func (d DataSizeUnit) String() string {
	switch d {
	case Bytes:
		return "b"
	case Kilobytes:
		return "kb"
	case Kibibytes:
		return "kib"
	case Megabytes:
		return "mb"
	case Mebibytes:
		return "mib"
	case Gigabytes:
		return "gb"
	case Gibibytes:
		return "gib"
	case Terabytes:
		return "tb"
	case Tebibytes:
		return "tib"
	case Petabytes:
		return "pb"
	case Pebibytes:
		return "pib"
	default:
		return "<invalid>"
	}
}

//goland:noinspection GoMixedReceiverTypes
func (d *DataSizeUnit) UnmarshalBinary(data []byte) error {
	unit, err := ToDataSizeUnit(string(data))
	if err != nil {
		return err
	}
	*d = unit
	return nil
}

// ToDataSizeUnit converts to a DataSizeUnit from its string representation.
func ToDataSizeUnit(str string) (DataSizeUnit, error) {
	switch str {
	case "b":
		return Bytes, nil
	case "kb":
		return Kilobytes, nil
	case "kib":
		return Kibibytes, nil
	case "mb":
		return Megabytes, nil
	case "mib":
		return Mebibytes, nil
	case "gb":
		return Gigabytes, nil
	case "gib":
		return Gibibytes, nil
	case "tb":
		return Terabytes, nil
	case "tib":
		return Tebibytes, nil
	case "pb":
		return Petabytes, nil
	case "pib":
		return Pebibytes, nil
	default:
		return Bytes, fmt.Errorf("unrecognized DataSize unit: `%s`", str)
	}
}
