package v1alpha

import (
	"context"
	"fmt"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/runtime"
)

// Const is a constant with a native type recognized by Nuggit.
//
// See nuggit.Type.
type Const struct {
	Type  nuggit.Type `json:"type,omitempty"`
	Value any         `json:"value,omitempty"`
}

func (x *Const) Assign(t nuggit.Type, v any) {
	x.Type = t
	x.Value = v
}

func (x *Const) CopyTo(dst *Const) error {
	if dst.Type != nuggit.TypeUndefined && dst.Type != x.Type {
		return fmt.Errorf("type mismatch")
	}
	dst.Assign(x.Type, x.Value)
	return nil
}

func (x *Const) Bind(edges []runtime.Edge) error {
	if len(edges) != 0 {
		return fmt.Errorf("Const is a leaf node")
	}
	return nil
}

func (x *Const) Run(context.Context) (any, error) {
	return x.Value, nil
}
