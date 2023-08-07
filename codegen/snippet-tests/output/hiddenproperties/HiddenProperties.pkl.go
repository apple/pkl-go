// Code generated from Pkl module `HiddenProperties`. DO NOT EDIT.
package hiddenproperties

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type HiddenProperties struct {
	PropC string `pkl:"propC"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a HiddenProperties
func LoadFromPath(ctx context.Context, path string) (ret *HiddenProperties, err error) {
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a HiddenProperties
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (*HiddenProperties, error) {
	var ret HiddenProperties
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}
