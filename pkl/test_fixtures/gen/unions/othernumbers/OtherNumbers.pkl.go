// Code generated from Pkl module `unions`. DO NOT EDIT.
package othernumbers

import (
	"encoding"
	"fmt"
)

type OtherNumbers string

const (
	N一 OtherNumbers = "一"
	N二 OtherNumbers = "二"
	N三 OtherNumbers = "三"
)

// String returns the string representation of OtherNumbers
func (rcv OtherNumbers) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(OtherNumbers)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for OtherNumbers.
func (rcv *OtherNumbers) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "一":
		*rcv = N一
	case "二":
		*rcv = N二
	case "三":
		*rcv = N三
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid OtherNumbers`, str)
	}
	return nil
}
