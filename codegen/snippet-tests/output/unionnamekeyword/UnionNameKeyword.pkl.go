// Code generated from Pkl module `UnionNameKeyword`. DO NOT EDIT.
package unionnamekeyword

import (
	"context"

	"github.com/apple/pkl-go/codegen/snippet-tests/output/unionnamekeyword/_type"
	"github.com/apple/pkl-go/pkl"
)

type UnionNameKeyword struct {
	Type _type.Type `pkl:"type"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a UnionNameKeyword
func LoadFromPath(ctx context.Context, path string) (ret UnionNameKeyword, err error) {
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a UnionNameKeyword
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (UnionNameKeyword, error) {
	var ret UnionNameKeyword
	err := evaluator.EvaluateModule(ctx, source, &ret)
	return ret, err
}
