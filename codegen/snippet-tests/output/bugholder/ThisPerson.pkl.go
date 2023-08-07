// Code generated from Pkl module `org.foo.BugHolder`. DO NOT EDIT.
package bugholder

type ThisPerson interface {
	Person

	GetMyself() ThisPerson

	GetSomeoneElse() Person
}

var _ ThisPerson = (*ThisPersonImpl)(nil)

type ThisPersonImpl struct {
	*PersonImpl

	Myself ThisPerson `pkl:"myself"`

	SomeoneElse Person `pkl:"someoneElse"`
}

func (rcv *ThisPersonImpl) GetMyself() ThisPerson {
	return rcv.Myself
}

func (rcv *ThisPersonImpl) GetSomeoneElse() Person {
	return rcv.SomeoneElse
}
