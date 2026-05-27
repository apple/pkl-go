// Code generated from Pkl module `reference`. DO NOT EDIT.
package reference

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Reference struct {
	Res0 pkl.Reference[D] `pkl:"res0"`

	Res1 pkl.Reference[D] `pkl:"res1"`

	Res2 pkl.Reference[D] `pkl:"res2"`

	Res3 pkl.Reference[D] `pkl:"res3"`

	Res4 pkl.Reference[D] `pkl:"res4"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Reference
func LoadFromPath(ctx context.Context, path string) (ret Reference, err error) {
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Reference
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (Reference, error) {
	var ret Reference
	err := evaluator.EvaluateModule(ctx, source, &ret)
	return ret, err
}
