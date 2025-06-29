// Code generated from Pkl module `ExtendingOpenClass`. DO NOT EDIT.
package extendopenclass

type IMyOpenClass interface {
	GetMyStr() string
}

var _ IMyOpenClass = MyOpenClass{}

type MyOpenClass struct {
	MyStr string `pkl:"myStr"`
}

func (rcv MyOpenClass) GetMyStr() string {
	return rcv.MyStr
}
