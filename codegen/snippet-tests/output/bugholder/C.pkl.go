// Code generated from Pkl module `org.foo.BugHolder`. DO NOT EDIT.
package bugholder

type C interface {
	B

	GetC() string
}

var _ C = CImpl{}

type CImpl struct {
	BImpl

	C string `pkl:"c"`
}

func (rcv CImpl) GetC() string {
	return rcv.C
}
