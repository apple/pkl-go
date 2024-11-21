// Code generated from Pkl module `datasize`. DO NOT EDIT.
package datasize

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Datasize struct {
	Res1 *pkl.DataSize `pkl:"res1"`

	Res2 *pkl.DataSize `pkl:"res2"`

	Res3 *pkl.DataSize `pkl:"res3"`

	Res4 *pkl.DataSize `pkl:"res4"`

	Res5 *pkl.DataSize `pkl:"res5"`

	Res6 *pkl.DataSize `pkl:"res6"`

	Res7 *pkl.DataSize `pkl:"res7"`

	Res8 *pkl.DataSize `pkl:"res8"`

	Res9 *pkl.DataSize `pkl:"res9"`

	Res10 *pkl.DataSize `pkl:"res10"`

	Res11 *pkl.DataSize `pkl:"res11"`

	Res12 pkl.DataSizeUnit `pkl:"res12"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Datasize
func LoadFromPath(ctx context.Context, path string) (ret Datasize, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return Datasize{}, err
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Datasize
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (Datasize, error) {
	var ret Datasize
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return Datasize{}, err
	}
	return ret, nil
}
