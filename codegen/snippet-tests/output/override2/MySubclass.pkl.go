// Code generated from Pkl module `Override2`. DO NOT EDIT.
package override2

type IMySubclass interface {
	IOverride2

	GetFoo() string
}

var _ IMySubclass = MySubclass{}

type MySubclass struct {
	Override2

	// Different doc comments
	Foo string `pkl:"foo"`
}

// Different doc comments
func (rcv MySubclass) GetFoo() string {
	return rcv.Foo
}
