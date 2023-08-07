// Code generated from Pkl module `ExtendingOpenClass`. DO NOT EDIT.
package extendopenclass

import "github.com/apple/pkl-go/codegen/snippet-tests/output/support/lib3"

type MyClass2 interface {
	lib3.GoGoGo

	GetMyBoolean() bool
}

var _ MyClass2 = (*MyClass2Impl)(nil)

type MyClass2Impl struct {
	*lib3.GoGoGoImpl

	MyBoolean bool `pkl:"myBoolean"`
}

func (rcv *MyClass2Impl) GetMyBoolean() bool {
	return rcv.MyBoolean
}
