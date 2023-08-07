// Code generated from Pkl module `duration`. DO NOT EDIT.
package duration

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Duration struct {
	Res1 *pkl.Duration `pkl:"res1"`

	Res2 *pkl.Duration `pkl:"res2"`

	Res3 *pkl.Duration `pkl:"res3"`

	Res4 *pkl.Duration `pkl:"res4"`

	Res5 *pkl.Duration `pkl:"res5"`

	Res6 *pkl.Duration `pkl:"res6"`

	Res7 *pkl.Duration `pkl:"res7"`

	Res8 pkl.DurationUnit `pkl:"res8"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Duration
func LoadFromPath(ctx context.Context, path string) (ret *Duration, err error) {
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Duration
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (*Duration, error) {
	var ret Duration
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}
