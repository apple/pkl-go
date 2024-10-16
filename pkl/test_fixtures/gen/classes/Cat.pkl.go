// Code generated from Pkl module `classes`. DO NOT EDIT.
package classes

type ICat interface {
	IAnimal

	GetMeows() bool
}

var _ ICat = Cat{}

type Cat struct {
	Meows bool `pkl:"meows"`

	Name string `pkl:"name"`
}

func (rcv Cat) GetMeows() bool {
	return rcv.Meows
}

func (rcv Cat) GetName() string {
	return rcv.Name
}
