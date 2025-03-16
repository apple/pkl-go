// Code generated from Pkl module `classes`. DO NOT EDIT.
package classes

type Dog interface {
	Animal

	GetBarks() bool

	GetBreed() string
}

var _ Dog = DogImpl{}

type DogImpl struct {
	Barks bool `pkl:"barks"`

	Breed string `pkl:"breed"`

	Name string `pkl:"name"`
}

func (rcv DogImpl) GetBarks() bool {
	return rcv.Barks
}

func (rcv DogImpl) GetBreed() string {
	return rcv.Breed
}

func (rcv DogImpl) GetName() string {
	return rcv.Name
}
