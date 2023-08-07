// Code generated from Pkl module `org.foo.BugHolder`. DO NOT EDIT.
package bugholder

type Person interface {
	Being

	GetBike() *Bike

	GetFirstName() *uint16

	GetLastName() map[string]*uint32

	GetThings() map[int]struct{}
}

var _ Person = (*PersonImpl)(nil)

// A Person!
type PersonImpl struct {
	IsAlive bool `pkl:"isAlive"`

	Bike *Bike `pkl:"bike"`

	// The person's first name
	FirstName *uint16 `pkl:"firstName"`

	// The person's last name
	LastName map[string]*uint32 `pkl:"lastName"`

	Things map[int]struct{} `pkl:"things"`
}

func (rcv *PersonImpl) GetIsAlive() bool {
	return rcv.IsAlive
}

func (rcv *PersonImpl) GetBike() *Bike {
	return rcv.Bike
}

// The person's first name
func (rcv *PersonImpl) GetFirstName() *uint16 {
	return rcv.FirstName
}

// The person's last name
func (rcv *PersonImpl) GetLastName() map[string]*uint32 {
	return rcv.LastName
}

func (rcv *PersonImpl) GetThings() map[int]struct{} {
	return rcv.Things
}
