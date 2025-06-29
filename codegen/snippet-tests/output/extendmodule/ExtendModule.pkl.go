// Code generated from Pkl module `ExtendModule`. DO NOT EDIT.
package extendmodule

import (
	"context"

	"github.com/apple/pkl-go/codegen/snippet-tests/output/support/openmodule"
	"github.com/apple/pkl-go/pkl"
)

type IExtendModule interface {
	openmodule.IMyModule

	GetBar() string
}

var _ IExtendModule = ExtendModule{}

type ExtendModule struct {
	openmodule.MyModule

	Bar string `pkl:"bar"`
}

func (rcv ExtendModule) GetBar() string {
	return rcv.Bar
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a ExtendModule
func LoadFromPath(ctx context.Context, path string) (ret ExtendModule, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return ExtendModule{}, err
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a ExtendModule
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (ExtendModule, error) {
	var ret ExtendModule
	if err := evaluator.EvaluateModule(ctx, source, &ret); err != nil {
		return ExtendModule{}, err
	}
	return ret, nil
}
