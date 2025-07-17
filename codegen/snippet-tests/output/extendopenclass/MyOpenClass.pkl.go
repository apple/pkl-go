// Code generated from Pkl module `ExtendingOpenClass`. DO NOT EDIT.
package extendopenclass

type MyOpenClass interface {
	GetMyStr() string
}

var _ MyOpenClass = MyOpenClassImpl{}

type MyOpenClassImpl struct {
	MyStr string `pkl:"myStr"`
}

func (rcv MyOpenClassImpl) GetMyStr() string {
	return rcv.MyStr
}
