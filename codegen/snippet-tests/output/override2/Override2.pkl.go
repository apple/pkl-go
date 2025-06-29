// Code generated from Pkl module `Override2`. DO NOT EDIT.
package override2

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type IOverride2 interface {
	GetFoo() string
}

var _ IOverride2 = Override2{}

type Override2 struct {
	// Doc comments
	Foo string `pkl:"foo"`
}

// Doc comments
func (rcv Override2) GetFoo() string {
	return rcv.Foo
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Override2
func LoadFromPath(ctx context.Context, path string) (ret Override2, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return Override2{}, err
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Override2
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (Override2, error) {
	var ret Override2
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return Override2{}, err
	}
	return ret, nil
}
