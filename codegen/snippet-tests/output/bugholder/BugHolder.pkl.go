// Code generated from Pkl module `org.foo.BugHolder`. DO NOT EDIT.
package bugholder

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type BugHolder struct {
	Bug Bug `pkl:"bug"`

	N蚊子 Bug `pkl:"蚊子"`

	ThisPerson IThisPerson `pkl:"thisPerson"`

	D ID `pkl:"d"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a BugHolder
func LoadFromPath(ctx context.Context, path string) (ret BugHolder, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return BugHolder{}, err
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a BugHolder
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (BugHolder, error) {
	var ret BugHolder
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return BugHolder{}, err
	}
	return ret, nil
}
