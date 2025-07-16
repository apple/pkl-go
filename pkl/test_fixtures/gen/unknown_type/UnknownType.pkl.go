// Code generated from Pkl module `unknown_type`. DO NOT EDIT.
package unknowntype

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type UnknownType struct {
	Res any `pkl:"res"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a UnknownType
func LoadFromPath(ctx context.Context, path string) (ret UnknownType, err error) {
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a UnknownType
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (UnknownType, error) {
	var ret UnknownType
	err := evaluator.EvaluateModule(ctx, source, &ret)
	return ret, err
}
