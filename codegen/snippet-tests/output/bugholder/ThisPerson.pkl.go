// Code generated from Pkl module `org.foo.BugHolder`. DO NOT EDIT.
package bugholder

type IThisPerson interface {
	IPerson

	GetMyself() *IThisPerson

	GetSomeoneElse() IPerson
}

var _ IThisPerson = ThisPerson{}

type ThisPerson struct {
	Person

	Myself *IThisPerson `pkl:"myself"`

	SomeoneElse IPerson `pkl:"someoneElse"`
}

func (rcv ThisPerson) GetMyself() *IThisPerson {
	return rcv.Myself
}

func (rcv ThisPerson) GetSomeoneElse() IPerson {
	return rcv.SomeoneElse
}
