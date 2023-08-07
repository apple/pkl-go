// Code generated from Pkl module `union`. DO NOT EDIT.
package city

import (
	"encoding"
	"fmt"
)

// City; e.g. where people live
type City string

const (
	SanFrancisco City = "San Francisco"
	London       City = "London"
	N上海          City = "上海"
)

// String returns the string representation of City
func (rcv City) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(City)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for City.
func (rcv *City) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "San Francisco":
		*rcv = SanFrancisco
	case "London":
		*rcv = London
	case "上海":
		*rcv = N上海
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid City`, str)
	}
	return nil
}
