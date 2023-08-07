// Code generated from Pkl module `Override2`. DO NOT EDIT.
package override2

type MySubclass interface {
	Override2

	GetFoo() string
}

var _ MySubclass = (*MySubclassImpl)(nil)

type MySubclassImpl struct {
	*Override2Impl

	// Different doc comments
	Foo string `pkl:"foo"`
}

// Different doc comments
func (rcv *MySubclassImpl) GetFoo() string {
	return rcv.Foo
}
