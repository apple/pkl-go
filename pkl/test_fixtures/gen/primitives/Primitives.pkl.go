// Code generated from Pkl module `primitives`. DO NOT EDIT.
package primitives

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Primitives struct {
	Res11 any `pkl:"res11"`

	Res0 string `pkl:"res0"`

	Res1 int `pkl:"res1"`

	Res5 uint `pkl:"res5"`

	Res9 float64 `pkl:"res9"`

	Res12 float64 `pkl:"res12"`

	Res13 float64 `pkl:"res13"`

	Res4 int32 `pkl:"res4"`

	Res8 uint32 `pkl:"res8"`

	Res3 int16 `pkl:"res3"`

	Res7 uint16 `pkl:"res7"`

	Res2 int8 `pkl:"res2"`

	Res6 uint8 `pkl:"res6"`

	Res10 bool `pkl:"res10"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Primitives
func LoadFromPath(ctx context.Context, path string) (ret Primitives, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return ret, err
	}
	defer func() {
		cerr := evaluator.Close()
		if err == nil {
			err = cerr
		}
	}()
	ret, err = Load(ctx, evaluator, pkl.FileSource(path))
	return ret, err
}

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Primitives
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (Primitives, error) {
	var ret Primitives
	err := evaluator.EvaluateModule(ctx, source, &ret)
	return ret, err
}
