// Code generated from Pkl module `ExtendsAbstractClass`. DO NOT EDIT.
package extendabstractclass

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type ExtendsAbstractClass struct {
	A IA `pkl:"a"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a ExtendsAbstractClass
func LoadFromPath(ctx context.Context, path string) (ret ExtendsAbstractClass, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return ExtendsAbstractClass{}, err
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a ExtendsAbstractClass
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (ExtendsAbstractClass, error) {
	var ret ExtendsAbstractClass
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return ExtendsAbstractClass{}, err
	}
	return ret, nil
}
