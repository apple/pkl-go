// Code generated from Pkl module `EmptyOpenModule`. DO NOT EDIT.
package emptyopenmodule

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type EmptyOpenModule interface {
}

var _ EmptyOpenModule = EmptyOpenModuleImpl{}

type EmptyOpenModuleImpl struct {
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a EmptyOpenModuleImpl
func LoadFromPath(ctx context.Context, path string) (ret EmptyOpenModuleImpl, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return EmptyOpenModuleImpl{}, err
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a EmptyOpenModuleImpl
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (EmptyOpenModuleImpl, error) {
	var ret EmptyOpenModuleImpl
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return EmptyOpenModuleImpl{}, err
	}
	return ret, nil
}
