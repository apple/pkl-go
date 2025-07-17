// Code generated from Pkl module `collections`. DO NOT EDIT.
package collections

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Collections struct {
	Res1 []int `pkl:"res1"`

	Res2 []int `pkl:"res2"`

	Res3 [][]int `pkl:"res3"`

	Res4 [][]int `pkl:"res4"`

	Res5 map[int]bool `pkl:"res5"`

	Res6 map[int]map[int]bool `pkl:"res6"`

	Res7 map[int]bool `pkl:"res7"`

	Res8 map[int]map[int]bool `pkl:"res8"`

	Res9 map[string]struct{} `pkl:"res9"`

	Res10 map[int8]struct{} `pkl:"res10"`

	Res11 *pkl.Pair[int, float64] `pkl:"res11"`

	Res12 *pkl.Pair[any, any] `pkl:"res12"`

	Res13 *pkl.Pair[int, *int] `pkl:"res13"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Collections
func LoadFromPath(ctx context.Context, path string) (ret Collections, err error) {
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Collections
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (Collections, error) {
	var ret Collections
	err := evaluator.EvaluateModule(ctx, source, &ret)
	return ret, err
}
