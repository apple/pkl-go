// Code generated from Pkl module `union`. DO NOT EDIT.
package union

import (
	"context"

	"github.com/apple/pkl-go/codegen/snippet-tests/output/union/accountdisposition"
	"github.com/apple/pkl-go/codegen/snippet-tests/output/union/city"
	"github.com/apple/pkl-go/codegen/snippet-tests/output/union/county"
	"github.com/apple/pkl-go/codegen/snippet-tests/output/union/noodles"
	"github.com/apple/pkl-go/pkl"
)

type Union struct {
	// A city
	City city.City `pkl:"city"`

	// County
	County county.County `pkl:"county"`

	// Noodles
	Noodle noodles.Noodles `pkl:"noodle"`

	// Account disposition
	Disposition accountdisposition.AccountDisposition `pkl:"disposition"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Union
func LoadFromPath(ctx context.Context, path string) (ret *Union, err error) {
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Union
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (*Union, error) {
	var ret Union
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}
