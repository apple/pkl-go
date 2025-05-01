// Code generated from Pkl module `ExtendingOpenClass`. DO NOT EDIT.
package extendopenclass

import "github.com/apple/pkl-go/codegen/snippet-tests/output/support/lib3"

type IMyClass2 interface {
	lib3.IGoGoGo

	GetMyBoolean() bool
}

var _ IMyClass2 = MyClass2{}

type MyClass2 struct {
	lib3.GoGoGo

	MyBoolean bool `pkl:"myBoolean"`
}

func (rcv MyClass2) GetMyBoolean() bool {
	return rcv.MyBoolean
}
