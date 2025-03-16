// Code generated from Pkl module `ExtendsAbstractClass`. DO NOT EDIT.
package extendabstractclass

type B interface {
	A

	GetC() string
}

var _ B = BImpl{}

type BImpl struct {
	B string `pkl:"b"`

	C string `pkl:"c"`
}

func (rcv BImpl) GetB() string {
	return rcv.B
}

func (rcv BImpl) GetC() string {
	return rcv.C
}
