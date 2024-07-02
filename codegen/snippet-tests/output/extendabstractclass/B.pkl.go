// Code generated from Pkl module `ExtendsAbstractClass`. DO NOT EDIT.
package extendabstractclass

type IB interface {
	IA

	GetC() string
}

var _ IB = B{}

type B struct {
	B string `pkl:"b"`

	C string `pkl:"c"`
}

func (rcv B) GetB() string {
	return rcv.B
}

func (rcv B) GetC() string {
	return rcv.C
}
