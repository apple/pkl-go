// Code generated from Pkl module `override`. DO NOT EDIT.
package override

type IBar interface {
	IFoo
}

var _ IBar = Bar{}

type Bar struct {
	MyProp string `pkl:"myProp"`
}

func (rcv Bar) GetMyProp() string {
	return rcv.MyProp
}
