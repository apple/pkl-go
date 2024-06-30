// Code generated from Pkl module `MyModule`. DO NOT EDIT.
package openmodule

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type MyModule interface {
	GetFoo() string
}

var _ MyModule = MyModuleImpl{}

type MyModuleImpl struct {
	Foo string `pkl:"foo"`
}

func (rcv MyModuleImpl) GetFoo() string {
	return rcv.Foo
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a MyModuleImpl
func LoadFromPath(ctx context.Context, path string) (ret MyModuleImpl, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return MyModuleImpl{}, err
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a MyModuleImpl
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (MyModuleImpl, error) {
	var ret MyModuleImpl
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return MyModuleImpl{}, err
	}
	return ret, nil
}
