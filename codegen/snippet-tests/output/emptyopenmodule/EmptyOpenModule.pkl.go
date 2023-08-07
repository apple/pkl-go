// Code generated from Pkl module `EmptyOpenModule`. DO NOT EDIT.
package emptyopenmodule

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type EmptyOpenModule interface {
}

var _ EmptyOpenModule = (*EmptyOpenModuleImpl)(nil)

type EmptyOpenModuleImpl struct {
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a EmptyOpenModule
func LoadFromPath(ctx context.Context, path string) (ret EmptyOpenModule, err error) {
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a EmptyOpenModule
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (EmptyOpenModule, error) {
	var ret EmptyOpenModuleImpl
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}
