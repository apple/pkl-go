// Code generated from Pkl module `classes`. DO NOT EDIT.
package classes

type Cat interface {
	Animal

	GetMeows() bool
}

var _ Cat = (*CatImpl)(nil)

type CatImpl struct {
	Meows bool `pkl:"meows"`

	Name string `pkl:"name"`
}

func (rcv *CatImpl) GetMeows() bool {
	return rcv.Meows
}

func (rcv *CatImpl) GetName() string {
	return rcv.Name
}
