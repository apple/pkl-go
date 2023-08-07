// Code generated from Pkl module `org.foo.BugHolder`. DO NOT EDIT.
package bugkindtwo

import (
	"encoding"
	"fmt"
)

type BugKindTwo string

const (
	Butterfly BugKindTwo = "butterfly"
	Beetle    BugKindTwo = `beetle"`
	BeetleOne BugKindTwo = "beetle one"
)

// String returns the string representation of BugKindTwo
func (rcv BugKindTwo) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(BugKindTwo)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for BugKindTwo.
func (rcv *BugKindTwo) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "butterfly":
		*rcv = Butterfly
	case `beetle"`:
		*rcv = Beetle
	case "beetle one":
		*rcv = BeetleOne
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid BugKindTwo`, str)
	}
	return nil
}
