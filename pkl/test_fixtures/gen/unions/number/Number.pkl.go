// Code generated from Pkl module `unions`. DO NOT EDIT.
package number

import (
	"encoding"
	"fmt"
)

type Number string

const (
	One   Number = "one"
	Two   Number = "two"
	Three Number = "three"
)

// String returns the string representation of Number
func (rcv Number) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(Number)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for Number.
func (rcv *Number) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "one":
		*rcv = One
	case "two":
		*rcv = Two
	case "three":
		*rcv = Three
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid Number`, str)
	}
	return nil
}
