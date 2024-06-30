// Code generated from Pkl module `ExtendingOpenClass`. DO NOT EDIT.
package extendopenclass

type IMyClass interface {
	IMyOpenClass

	GetMyBoolean() bool
}

var _ IMyClass = MyClass{}

type MyClass struct {
	MyOpenClass

	MyBoolean bool `pkl:"myBoolean"`
}

func (rcv MyClass) GetMyBoolean() bool {
	return rcv.MyBoolean
}
