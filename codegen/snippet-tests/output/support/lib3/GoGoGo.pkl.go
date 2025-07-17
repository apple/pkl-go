// Code generated from Pkl module `lib3`. DO NOT EDIT.
package lib3

type GoGoGo interface {
	GetDuck() string
}

var _ GoGoGo = GoGoGoImpl{}

type GoGoGoImpl struct {
	Duck string `pkl:"duck"`
}

func (rcv GoGoGoImpl) GetDuck() string {
	return rcv.Duck
}
