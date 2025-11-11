// Code generated from Pkl module `classes`. DO NOT EDIT.
package classes

type Cat interface {
	Animal

	GetMeows() bool
}

var _ Cat = CatImpl{}

type CatImpl struct {
	Name  string `pkl:"name"`
	Meows bool   `pkl:"meows"`
}

func (rcv CatImpl) GetMeows() bool {
	return rcv.Meows
}

func (rcv CatImpl) GetName() string {
	return rcv.Name
}
