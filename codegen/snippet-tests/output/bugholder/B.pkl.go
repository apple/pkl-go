// Code generated from Pkl module `org.foo.BugHolder`. DO NOT EDIT.
package bugholder

type IB interface {
	IA

	GetB() string
}

var _ IB = B{}

type B struct {
	B string `pkl:"b"`

	A string `pkl:"a"`
}

func (rcv B) GetB() string {
	return rcv.B
}

func (rcv B) GetA() string {
	return rcv.A
}
