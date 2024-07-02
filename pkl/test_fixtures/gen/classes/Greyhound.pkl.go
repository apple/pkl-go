// Code generated from Pkl module `classes`. DO NOT EDIT.
package classes

type IGreyhound interface {
	IDog

	GetCanRoach() bool
}

var _ IGreyhound = Greyhound{}

type Greyhound struct {
	Dog

	CanRoach bool `pkl:"canRoach"`
}

func (rcv Greyhound) GetCanRoach() bool {
	return rcv.CanRoach
}
