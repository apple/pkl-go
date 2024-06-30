// Code generated from Pkl module `ModuleUsingLib`. DO NOT EDIT.
package moduleusinglib

import (
	"context"

	"github.com/apple/pkl-go/codegen/snippet-tests/output/support/lib"
	"github.com/apple/pkl-go/codegen/snippet-tests/output/support/lib/myenum"
	"github.com/apple/pkl-go/codegen/snippet-tests/output/support/lib2/cities"
	"github.com/apple/pkl-go/pkl"
)

type ModuleUsingLib struct {
	Res []lib.MyClass `pkl:"res"`

	Res2 myenum.MyEnum `pkl:"res2"`

	Res3 string `pkl:"res3"`

	Res4 cities.Cities `pkl:"res4"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a ModuleUsingLib
func LoadFromPath(ctx context.Context, path string) (ret ModuleUsingLib, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return ModuleUsingLib{}, err
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a ModuleUsingLib
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (ModuleUsingLib, error) {
	var ret ModuleUsingLib
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return ModuleUsingLib{}, err
	}
	return ret, nil
}
