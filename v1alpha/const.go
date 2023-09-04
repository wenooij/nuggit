package v1alpha

import (
	"fmt"

	"github.com/wenooij/nuggit"
)

// Const is a constant with a native type recognized by Nuggit.
//
// See nuggit.Type.
//
// Deprecated: All primitives will be created directly.
type Const struct {
	Type  nuggit.Type `json:"type,omitempty"`
	Value any         `json:"value,omitempty"`
}

func (x *Const) CopyTo(v *Var) error {
	if x.Type != nuggit.TypeUndefined && x.Type != v.Type {
		return fmt.Errorf("type mismatch")
	}
	v.Value = x.Value
	return nil
}
