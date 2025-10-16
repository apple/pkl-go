// Code generated from Pkl module `types`. DO NOT EDIT.
package types

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Types struct {
	StringClass pkl.Class `pkl:"stringClass"`

	BaseModuleClass pkl.Class `pkl:"baseModuleClass"`

	Uint8TypeAlias pkl.TypeAlias `pkl:"uint8TypeAlias"`

	FooClass pkl.Class `pkl:"fooClass"`

	BarTypeAlias pkl.TypeAlias `pkl:"barTypeAlias"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Types
func LoadFromPath(ctx context.Context, path string) (ret Types, err error) {
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Types
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (Types, error) {
	var ret Types
	err := evaluator.EvaluateModule(ctx, source, &ret)
	return ret, err
}
