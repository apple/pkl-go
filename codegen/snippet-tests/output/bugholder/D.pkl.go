// Code generated from Pkl module `org.foo.BugHolder`. DO NOT EDIT.
package bugholder

type D interface {
	C

	GetD() string
}

var _ D = DImpl{}

type DImpl struct {
	CImpl

	D string `pkl:"d"`
}

func (rcv DImpl) GetD() string {
	return rcv.D
}
