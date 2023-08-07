// Code generated from Pkl module `org.foo.BugHolder`. DO NOT EDIT.
package bugholder

type B interface {
	A

	GetB() string
}

var _ B = (*BImpl)(nil)

type BImpl struct {
	B string `pkl:"b"`

	A string `pkl:"a"`
}

func (rcv *BImpl) GetB() string {
	return rcv.B
}

func (rcv *BImpl) GetA() string {
	return rcv.A
}
