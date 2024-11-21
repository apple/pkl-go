// Code generated from Pkl module `any`. DO NOT EDIT.
package any

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Any struct {
	Res1 any `pkl:"res1"`

	Res2 any `pkl:"res2"`

	Res3 any `pkl:"res3"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Any
func LoadFromPath(ctx context.Context, path string) (ret Any, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return Any{}, err
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Any
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (Any, error) {
	var ret Any
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return Any{}, err
	}
	return ret, nil
}
