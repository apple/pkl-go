// Code generated from Pkl module `lib4`. DO NOT EDIT.
package lib4

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Lib4 struct {
	Bar string `pkl:"bar"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Lib4
func LoadFromPath(ctx context.Context, path string) (ret *Lib4, err error) {
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Lib4
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (*Lib4, error) {
	var ret Lib4
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}
