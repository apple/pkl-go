// Code generated from Pkl module `union`. DO NOT EDIT.
package noodles

import (
	"encoding"
	"fmt"
)

// Noodles
type Noodles string

const (
	N拉面   Noodles = "拉面"
	N刀切面  Noodles = "刀切面"
	N面线   Noodles = "面线"
	N意大利面 Noodles = "意大利面"
)

// String returns the string representation of Noodles
func (rcv Noodles) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(Noodles)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for Noodles.
func (rcv *Noodles) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "拉面":
		*rcv = N拉面
	case "刀切面":
		*rcv = N刀切面
	case "面线":
		*rcv = N面线
	case "意大利面":
		*rcv = N意大利面
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid Noodles`, str)
	}
	return nil
}
