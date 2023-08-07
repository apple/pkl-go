// Code generated from Pkl module `union`. DO NOT EDIT.
package county

import (
	"encoding"
	"fmt"
)

// Locale that contains cities and towns
type County string

const (
	SanFrancisco County = "San Francisco"
	SanMateo     County = "San Mateo"
	Yolo         County = "Yolo"
)

// String returns the string representation of County
func (rcv County) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(County)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for County.
func (rcv *County) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "San Francisco":
		*rcv = SanFrancisco
	case "San Mateo":
		*rcv = SanMateo
	case "Yolo":
		*rcv = Yolo
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid County`, str)
	}
	return nil
}
