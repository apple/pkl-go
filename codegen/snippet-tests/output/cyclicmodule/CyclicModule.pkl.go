// Code generated from Pkl module `CyclicModule`. DO NOT EDIT.
package cyclicmodule

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type CyclicModule struct {
	Thing Cyclic `pkl:"thing"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a CyclicModule
func LoadFromPath(ctx context.Context, path string) (ret CyclicModule, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return CyclicModule{}, err
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a CyclicModule
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (CyclicModule, error) {
	var ret CyclicModule
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return CyclicModule{}, err
	}
	return ret, nil
}
