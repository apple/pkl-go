// Code generated from Pkl module `classes`. DO NOT EDIT.
package classes

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Classes struct {
	Animals []Animal `pkl:"animals"`

	NullableAnimals []*Animal `pkl:"nullableAnimals"`

	MyAnimal Animal `pkl:"myAnimal"`

	House House `pkl:"house"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Classes
func LoadFromPath(ctx context.Context, path string) (ret Classes, err error) {
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Classes
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (Classes, error) {
	var ret Classes
	err := evaluator.EvaluateModule(ctx, source, &ret)
	return ret, err
}
