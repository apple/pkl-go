// Code generated from Pkl module `ModuleType`. DO NOT EDIT.
package moduletype

import (
	"context"

	"github.com/apple/pkl-go/codegen/snippet-tests/output/support/lib4"
	"github.com/apple/pkl-go/pkl"
)

type ModuleType struct {
	MyStr string `pkl:"myStr"`

	Foo *ModuleType `pkl:"foo"`

	Lib lib4.MyLib4 `pkl:"lib"`

	FooAgain *ModuleType `pkl:"fooAgain"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a ModuleType
func LoadFromPath(ctx context.Context, path string) (ret ModuleType, err error) {
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a ModuleType
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (ModuleType, error) {
	var ret ModuleType
	err := evaluator.EvaluateModule(ctx, source, &ret)
	return ret, err
}
