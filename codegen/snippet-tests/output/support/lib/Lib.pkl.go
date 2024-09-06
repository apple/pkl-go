// Code generated from Pkl module `lib`. DO NOT EDIT.
package lib

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Lib struct {
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Lib
func LoadFromPath(ctx context.Context, path string) (ret Lib, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return Lib{}, err
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Lib
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (Lib, error) {
	var ret Lib
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return Lib{}, err
	}
	return ret, nil
}
