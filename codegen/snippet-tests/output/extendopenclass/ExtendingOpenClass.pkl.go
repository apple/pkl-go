// Code generated from Pkl module `ExtendingOpenClass`. DO NOT EDIT.
package extendopenclass

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type ExtendingOpenClass struct {
	Res1 IMyClass `pkl:"res1"`

	Res2 IMyClass2 `pkl:"res2"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a ExtendingOpenClass
func LoadFromPath(ctx context.Context, path string) (ret ExtendingOpenClass, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return ExtendingOpenClass{}, err
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a ExtendingOpenClass
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (ExtendingOpenClass, error) {
	var ret ExtendingOpenClass
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return ExtendingOpenClass{}, err
	}
	return ret, nil
}
