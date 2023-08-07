// Code generated from Pkl module `union`. DO NOT EDIT.
package accountdisposition

import (
	"encoding"
	"fmt"
)

type AccountDisposition string

const (
	Empty   AccountDisposition = ""
	Icloud3 AccountDisposition = "icloud3"
	Prod    AccountDisposition = "prod"
	Shared  AccountDisposition = "shared"
)

// String returns the string representation of AccountDisposition
func (rcv AccountDisposition) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(AccountDisposition)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for AccountDisposition.
func (rcv *AccountDisposition) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "":
		*rcv = Empty
	case "icloud3":
		*rcv = Icloud3
	case "prod":
		*rcv = Prod
	case "shared":
		*rcv = Shared
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid AccountDisposition`, str)
	}
	return nil
}
