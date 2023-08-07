// Code generated from Pkl module `ExtendingOpenClass`. DO NOT EDIT.
package extendopenclass

type MyClass interface {
	MyOpenClass

	GetMyBoolean() bool
}

var _ MyClass = (*MyClassImpl)(nil)

type MyClassImpl struct {
	*MyOpenClassImpl

	MyBoolean bool `pkl:"myBoolean"`
}

func (rcv *MyClassImpl) GetMyBoolean() bool {
	return rcv.MyBoolean
}
