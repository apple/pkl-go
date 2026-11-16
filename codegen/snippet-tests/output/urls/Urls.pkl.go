// Code generated from Pkl module `Urls`. DO NOT EDIT.
package urls

import (
	"context"
	"net/url"

	"github.com/apple/pkl-go/pkl"
)

type Urls struct {
	Url url.URL `pkl:"url"`

	OptionalUrl *url.URL `pkl:"optionalUrl"`

	Urls []url.URL `pkl:"urls"`

	UrlsByName map[string]url.URL `pkl:"urlsByName"`

	AliasedUrl url.URL `pkl:"aliasedUrl"`

	Uri url.URL `pkl:"uri"`

	OptionalUri *url.URL `pkl:"optionalUri"`

	Uris []url.URL `pkl:"uris"`

	Endpoint Endpoint `pkl:"endpoint"`
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
