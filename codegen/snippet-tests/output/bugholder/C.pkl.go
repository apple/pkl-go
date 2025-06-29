// Code generated from Pkl module `org.foo.BugHolder`. DO NOT EDIT.
package bugholder

type IC interface {
	IB

	GetC() string
}

var _ IC = C{}

type C struct {
	B

	C string `pkl:"c"`
}

func (rcv C) GetC() string {
	return rcv.C
}
