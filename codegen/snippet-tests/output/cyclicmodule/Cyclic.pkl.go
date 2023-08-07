// Code generated from Pkl module `CyclicModule`. DO NOT EDIT.
package cyclicmodule

type Cyclic struct {
	A string `pkl:"a"`

	B int `pkl:"b"`

	Myself *Cyclic `pkl:"myself"`
}
