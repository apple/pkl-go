// Code generated from Pkl module `org.foo.BugHolder`. DO NOT EDIT.
package bugkind

import (
	"encoding"
	"fmt"
)

type BugKind string

const (
	Butterfly BugKind = "butterfly"
	Beetle    BugKind = `beetle"`
	BeetleOne BugKind = "beetle one"
)

// String returns the string representation of BugKind
func (rcv BugKind) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(BugKind)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for BugKind.
func (rcv *BugKind) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "butterfly":
		*rcv = Butterfly
	case `beetle"`:
		*rcv = Beetle
	case "beetle one":
		*rcv = BeetleOne
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid BugKind`, str)
	}
	return nil
}
