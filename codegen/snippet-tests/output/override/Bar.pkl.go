// Code generated from Pkl module `override`. DO NOT EDIT.
package override

type Bar interface {
	Foo
}

var _ Bar = (*BarImpl)(nil)

type BarImpl struct {
	MyProp string `pkl:"myProp"`
}

func (rcv *BarImpl) GetMyProp() string {
	return rcv.MyProp
}
