// Code generated from Pkl module `lib3`. DO NOT EDIT.
package lib3

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Lib3 struct {
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Lib3
func LoadFromPath(ctx context.Context, path string) (ret Lib3, err error) {
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Lib3
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (Lib3, error) {
	var ret Lib3
	err := evaluator.EvaluateModule(ctx, source, &ret)
	return ret, err
}
