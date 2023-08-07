// Code generated from Pkl module `nullables`. DO NOT EDIT.
package nullables

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Nullables struct {
	Res0 *string `pkl:"res0"`

	Res1 *string `pkl:"res1"`

	Res2 *int `pkl:"res2"`

	Res3 *int `pkl:"res3"`

	Res4 *int8 `pkl:"res4"`

	Res5 *int8 `pkl:"res5"`

	Res6 *int16 `pkl:"res6"`

	Res7 *int16 `pkl:"res7"`

	Res8 *int32 `pkl:"res8"`

	Res9 *int32 `pkl:"res9"`

	Res10 *uint `pkl:"res10"`

	Res11 *uint `pkl:"res11"`

	Res12 *uint8 `pkl:"res12"`

	Res13 *uint8 `pkl:"res13"`

	Res14 *uint16 `pkl:"res14"`

	Res15 *uint16 `pkl:"res15"`

	Res16 *uint32 `pkl:"res16"`

	Res17 *uint32 `pkl:"res17"`

	Res18 *float64 `pkl:"res18"`

	Res19 *float64 `pkl:"res19"`

	Res20 *bool `pkl:"res20"`

	Res21 *bool `pkl:"res21"`

	Res22 *map[string]string `pkl:"res22"`

	Res23 *map[string]string `pkl:"res23"`

	Res25 *map[*string]*string `pkl:"res25"`

	Res26 *[]*int `pkl:"res26"`

	Res27 *[]*int `pkl:"res27"`

	Res28 *MyClass `pkl:"res28"`

	Res29 *MyClass `pkl:"res29"`

	Res30 *MyClass `pkl:"res30"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Nullables
func LoadFromPath(ctx context.Context, path string) (ret *Nullables, err error) {
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Nullables
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (*Nullables, error) {
	var ret Nullables
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}
