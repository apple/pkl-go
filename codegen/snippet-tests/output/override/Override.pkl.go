// Code generated from Pkl module `override`. DO NOT EDIT.
package override

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Override struct {
	Foo Foo `pkl:"foo"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Override
func LoadFromPath(ctx context.Context, path string) (ret *Override, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return nil, err
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Override
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (*Override, error) {
	var ret Override
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}
