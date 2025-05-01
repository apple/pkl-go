// Code generated from Pkl module `dynamic`. DO NOT EDIT.
package dynamic

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Dynamic struct {
	Res1 pkl.Object `pkl:"res1"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Dynamic
func LoadFromPath(ctx context.Context, path string) (ret Dynamic, err error) {
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Dynamic
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (Dynamic, error) {
	var ret Dynamic
	err := evaluator.EvaluateModule(ctx, source, &ret)
	return ret, err
}
