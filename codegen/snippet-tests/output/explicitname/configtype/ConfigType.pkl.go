// Code generated from Pkl module `ExplicitName`. DO NOT EDIT.
package configtype

import (
	"encoding"
	"fmt"
)

type ConfigType string

const (
	One ConfigType = "one"
	Two ConfigType = "two"
)

// String returns the string representation of ConfigType
func (rcv ConfigType) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(ConfigType)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for ConfigType.
func (rcv *ConfigType) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "one":
		*rcv = One
	case "two":
		*rcv = Two
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid ConfigType`, str)
	}
	return nil
}
