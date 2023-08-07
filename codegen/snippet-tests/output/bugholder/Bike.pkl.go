// Code generated from Pkl module `org.foo.BugHolder`. DO NOT EDIT.
package bugholder

type Bike struct {
	IsFixie bool `pkl:"isFixie"`

	// Wheels are the front and back wheels.
	//
	// There are typically two of them.
	Wheels []*Wheel `pkl:"wheels"`
}
