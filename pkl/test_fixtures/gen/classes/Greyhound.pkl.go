// Code generated from Pkl module `classes`. DO NOT EDIT.
package classes

type Greyhound interface {
	Dog

	GetCanRoach() bool
}

var _ Greyhound = GreyhoundImpl{}

type GreyhoundImpl struct {
	DogImpl

	CanRoach bool `pkl:"canRoach"`
}

func (rcv GreyhoundImpl) GetCanRoach() bool {
	return rcv.CanRoach
}
