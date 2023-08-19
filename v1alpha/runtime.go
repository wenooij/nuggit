package v1alpha

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wenooij/nuggit"
)

type Edge struct {
	nuggit.Edge
	Result any
}

type Binder interface {
	json.Unmarshaler
	Bind([]Edge) error
}

type Runner interface {
	Run(context.Context) (any, error)
}

type BindRunner interface {
	Binder
	Runner
}

func NewRunner(n nuggit.Node) (BindRunner, error) {
	var op BindRunner
	switch n.Op {
	case "Const":
		op = &Const{}
	default:
		return nil, fmt.Errorf("NewRunner: Runner is not defined for Op: %q: %q", n.Key, n.Op)
	}
	if err := op.(Binder).UnmarshalJSON(n.Data); err != nil {
		return nil, err
	}
	return op, nil
}
