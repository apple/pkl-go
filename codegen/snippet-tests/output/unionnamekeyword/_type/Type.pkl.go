// Code generated from Pkl module `UnionNameKeyword`. DO NOT EDIT.
package _type

import (
	"encoding"
	"fmt"
)

type Type string

const (
	One   Type = "one"
	Two   Type = "two"
	Three Type = "three"
)

// String returns the string representation of Type
func (rcv Type) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(Type)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for Type.
func (rcv *Type) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "one":
		*rcv = One
	case "two":
		*rcv = Two
	case "three":
		*rcv = Three
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid Type`, str)
	}
	return nil
}
