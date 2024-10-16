// Code generated from Pkl module `lib`. DO NOT EDIT.
package lib

type MyClass interface {
	GetThing() string
}

var _ MyClass = MyClassImpl{}

type MyClassImpl struct {
	Thing string `pkl:"thing"`
}

func (rcv MyClassImpl) GetThing() string {
	return rcv.Thing
}
