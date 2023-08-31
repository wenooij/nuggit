package v1alpha

import (
	"context"
	"fmt"

	"github.com/wenooij/nuggit"
)

// Vap is a variable.
type Var struct {
	Type    nuggit.Type `json:"type,omitempty"`
	Value   any         `json:"value,omitempty"`
	Default any         `json:"default,omitempty"`
}

func (v *Var) Run(ctx context.Context) (any, error) {
	if v.Value != nil {
		return v.Value, nil
	}
	if v.Default == nil {
		return nil, fmt.Errorf("variable is unbound and without default")
	}
	v.Value = v.Default
	return v.Value, nil
}
