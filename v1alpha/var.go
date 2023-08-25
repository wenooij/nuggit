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

func (v *Var) Bind(e runtime.Edge) error {
	switch res := e.Result.(type) {
	case nuggit.Type:
		return nil
	case *Const:
		return nil
	default:
		return fmt.Errorf("unexpected type: %T", res)
	}
}

func (v *Var) Run(context.Context) (any, error) {
	if v.Value == nil {
		if v.Default == nil {
			return nil, fmt.Errorf("variable is unbound and without default")
		}
		v.Value = new(Const)
		v.Default.CopyTo(v.Value)
	}
	return v.Value, nil
}
