// Code generated from Pkl module `org.foo.BugHolder`. DO NOT EDIT.
package bugholder

import (
	"github.com/apple/pkl-go/codegen/snippet-tests/output/bugholder/bugkind"
	"github.com/apple/pkl-go/codegen/snippet-tests/output/bugholder/bugkindtwo"
	"github.com/apple/pkl-go/pkl"
)

type Bug struct {
	// The owner of this bug.
	Owner *Person `pkl:"owner"`

	// The age of this bug
	Age *int `pkl:"age"`

	// How long the bug holds its breath for
	HoldsBreathFor *pkl.Duration `pkl:"holdsBreathFor"`

	Size *pkl.DataSize `pkl:"size"`

	Kind bugkind.BugKind `pkl:"kind"`

	Kind2 bugkindtwo.BugKindTwo `pkl:"kind2"`

	Kind3 string `pkl:"kind3"`

	Kind4 string `pkl:"kind4"`

	BagOfStuff *pkl.Object `pkl:"bagOfStuff"`
}
