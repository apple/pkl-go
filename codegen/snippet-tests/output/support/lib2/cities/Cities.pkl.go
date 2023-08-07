// Code generated from Pkl module `lib2`. DO NOT EDIT.
package cities

import (
	"encoding"
	"fmt"
)

type Cities string

const (
	London       Cities = "London"
	SanFrancisco Cities = "San Francisco"
	LosAngeles   Cities = "Los Angeles"
)

// String returns the string representation of Cities
func (rcv Cities) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(Cities)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for Cities.
func (rcv *Cities) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "London":
		*rcv = London
	case "San Francisco":
		*rcv = SanFrancisco
	case "Los Angeles":
		*rcv = LosAngeles
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid Cities`, str)
	}
	return nil
}
