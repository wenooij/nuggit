package v1alpha

import (
	"context"
	"fmt"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/runtime"
)

// Vap is a variable.
type Var struct {
	Value   *Const `json:"value,omitempty"`
	Default *Const `json:"default,omitempty"`
}

func (v *Var) Bind(edges []runtime.Edge) error {
	for i, e := range edges {
		switch res := e.Result.(type) {
		case nuggit.Type:
		case *Const:
		default:
			return fmt.Errorf("unexpected input type at Edge[%d]: %T", i, res)
		}
	}
	if v.Value == nil {
		if v.Default == nil {
			return fmt.Errorf("variable is unbound and without default")
		}
		v.Value = new(Const)
	}
	return v.Default.CopyTo(v.Value)
}

func (v *Var) Run(context.Context) (any, error) {
	return v.Value, nil
}
