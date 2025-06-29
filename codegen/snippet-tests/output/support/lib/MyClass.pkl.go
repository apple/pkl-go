// Code generated from Pkl module `lib`. DO NOT EDIT.
package lib

type IMyClass interface {
	GetThing() string
}

var _ IMyClass = MyClass{}

type MyClass struct {
	Thing string `pkl:"thing"`
}

func (rcv MyClass) GetThing() string {
	return rcv.Thing
}
