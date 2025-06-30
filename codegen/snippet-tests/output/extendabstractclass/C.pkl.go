// Code generated from Pkl module `ExtendsAbstractClass`. DO NOT EDIT.
package extendabstractclass

import (
	"github.com/apple/pkl-go/codegen/snippet-tests/output/support/lib2/cities"
	"github.com/apple/pkl-go/codegen/snippet-tests/output/support/lib3"
)

type C interface {
	B

	GetD() string
}

var _ C = (*CImpl)(nil)

type CImpl struct {
	B string `pkl:"b"`

	D string `pkl:"d"`

	E cities.Cities `pkl:"e"`

	C lib3.GoGoGo `pkl:"c"`
}

func (rcv *CImpl) GetB() string {
	return rcv.B
}

func (rcv *CImpl) GetD() string {
	return rcv.D
}

func (rcv *CImpl) GetE() cities.Cities {
	return rcv.E
}

func (rcv *CImpl) GetC() lib3.GoGoGo {
	return rcv.C
}
