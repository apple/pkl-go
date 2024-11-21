// Code generated from Pkl module `lib3`. DO NOT EDIT.
package lib3

type IGoGoGo interface {
	GetDuck() string
}

var _ IGoGoGo = GoGoGo{}

type GoGoGo struct {
	Duck string `pkl:"duck"`
}

func (rcv GoGoGo) GetDuck() string {
	return rcv.Duck
}
