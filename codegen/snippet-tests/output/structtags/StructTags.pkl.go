// Code generated from Pkl module `StructTags`. DO NOT EDIT.
package structtags

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type StructTags struct {
	Res string `pkl:"res" json:"res,omitempty"`

	Res2 string `pkl:"res2" json:"res2"`

	Res3 string `pkl:"res3" json:"-"`

	Res5 string `pkl:"res5" json:"myFoo5,omitempty"`

	Res6 string `pkl:"res6" json:"myFoo6"`

	Res7 string `pkl:"res7" yaml:",omitempty" bson:",omitempty"`

	Res8 string `pkl:"res8" yaml:"res8,omitempty" bson:"res8,omitempty"`

	Res9 string `pkl:"res9"`

	Res10 string `pkl:"res10"`

	Res11 string `pkl:"res11" json:"-,omitempty"`

	Res12 string `pkl:"res12" toml:"res12,multiline"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a StructTags
func LoadFromPath(ctx context.Context, path string) (ret StructTags, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return StructTags{}, err
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a StructTags
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (StructTags, error) {
	var ret StructTags
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return StructTags{}, err
	}
	return ret, nil
}
