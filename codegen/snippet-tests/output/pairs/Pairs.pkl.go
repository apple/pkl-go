// Code generated from Pkl module `Pairs`. DO NOT EDIT.
package pairs

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Pairs interface {
	GetUntyped() pkl.Pair[any, any]

	GetOptional() *pkl.Pair[any, any]

	GetTyped() pkl.Pair[string, int]

	GetAliased() pkl.Pair[string, any]

	GetTypeArgAliased() pkl.Pair[string, int]
}

var _ Pairs = PairsImpl{}

type PairsImpl struct {
	Untyped pkl.Pair[any, any] `pkl:"untyped"`

	Optional *pkl.Pair[any, any] `pkl:"optional"`

	Typed pkl.Pair[string, int] `pkl:"typed"`

	Aliased pkl.Pair[string, any] `pkl:"aliased"`

	TypeArgAliased pkl.Pair[string, int] `pkl:"typeArgAliased"`
}

func (rcv PairsImpl) GetUntyped() pkl.Pair[any, any] {
	return rcv.Untyped
}

func (rcv PairsImpl) GetOptional() *pkl.Pair[any, any] {
	return rcv.Optional
}

func (rcv PairsImpl) GetTyped() pkl.Pair[string, int] {
	return rcv.Typed
}

func (rcv PairsImpl) GetAliased() pkl.Pair[string, any] {
	return rcv.Aliased
}

func (rcv PairsImpl) GetTypeArgAliased() pkl.Pair[string, int] {
	return rcv.TypeArgAliased
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Pairs
func LoadFromPath(ctx context.Context, path string) (ret Pairs, err error) {
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Pairs
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (Pairs, error) {
	var ret PairsImpl
	err := evaluator.EvaluateModule(ctx, source, &ret)
	return ret, err
}
