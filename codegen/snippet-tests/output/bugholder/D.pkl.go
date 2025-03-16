// Code generated from Pkl module `org.foo.BugHolder`. DO NOT EDIT.
package bugholder

type ID interface {
	IC

	GetD() string
}

var _ ID = D{}

type D struct {
	C

	D string `pkl:"d"`
}

func (rcv D) GetD() string {
	return rcv.D
}
