// Code generated from Pkl module `urls`. DO NOT EDIT.
package urls

import (
	"context"
	"net/url"

	"github.com/apple/pkl-go/pkl"
)

type Urls struct {
	Res1 *url.URL `pkl:"res1"`

	Res2 *url.URL `pkl:"res2"`

	Res3 *url.URL `pkl:"res3"`

	Res4 *url.URL `pkl:"res4"`

	Res5 *url.URL `pkl:"res5"`

	Res6 []*url.URL `pkl:"res6"`

	Res7 map[string]*url.URL `pkl:"res7"`

	Res8 any `pkl:"res8"`

	Res9 pkl.Object `pkl:"res9"`

	Res10 *url.URL `pkl:"res10"`

	Res11 *url.URL `pkl:"res11"`

	Res12 []*url.URL `pkl:"res12"`

	Res13 pkl.Object `pkl:"res13"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Urls
func LoadFromPath(ctx context.Context, path string) (ret Urls, err error) {
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Urls
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (Urls, error) {
	var ret Urls
	err := evaluator.EvaluateModule(ctx, source, &ret)
	return ret, err
}
