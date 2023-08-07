// Code generated from Pkl module `lib`. DO NOT EDIT.
package myenum

import (
	"encoding"
	"fmt"
)

type MyEnum string

const (
	One   MyEnum = "one"
	Two   MyEnum = "two"
	Three MyEnum = "three"
)

// String returns the string representation of MyEnum
func (rcv MyEnum) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(MyEnum)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for MyEnum.
func (rcv *MyEnum) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "one":
		*rcv = One
	case "two":
		*rcv = Two
	case "three":
		*rcv = Three
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid MyEnum`, str)
	}
	return nil
}
