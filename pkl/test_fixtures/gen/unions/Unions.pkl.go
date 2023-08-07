// Code generated from Pkl module `unions`. DO NOT EDIT.
package unions

import (
	"context"

	"github.com/apple/pkl-go/pkl"
	"github.com/apple/pkl-go/pkl/test_fixtures/gen/unions/number"
	"github.com/apple/pkl-go/pkl/test_fixtures/gen/unions/othernumbers"
)

type Unions struct {
	Res1 number.Number `pkl:"res1"`

	Res2 othernumbers.OtherNumbers `pkl:"res2"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Unions
func LoadFromPath(ctx context.Context, path string) (ret *Unions, err error) {
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Unions
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (*Unions, error) {
	var ret Unions
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}
