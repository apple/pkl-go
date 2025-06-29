// Code generated from Pkl module `classes`. DO NOT EDIT.
package classes

type IDog interface {
	IAnimal

	GetBarks() bool

	GetBreed() string
}

var _ IDog = Dog{}

type Dog struct {
	Barks bool `pkl:"barks"`

	Breed string `pkl:"breed"`

	Name string `pkl:"name"`
}

func (rcv Dog) GetBarks() bool {
	return rcv.Barks
}

func (rcv Dog) GetBreed() string {
	return rcv.Breed
}

func (rcv Dog) GetName() string {
	return rcv.Name
}
