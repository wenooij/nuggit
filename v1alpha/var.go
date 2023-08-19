package v1alpha

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wenooij/nuggit"
)

type VarData struct {
	Value   *Const `json:"value,omitempty"`
	Default *Const `json:"default,omitempty"`
}

// Vap is a variable.
type Var struct {
	VarData `json:",omitempty"`
}

func (v *Var) Bind(edges []Edge) error {
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

func (v *Var) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		v.Value = new(Const)
		v.Value.Assign(nuggit.TypeString, s)
		return nil
	}
	// TODO(wes): Support other types.
	var vd VarData
	if err := json.Unmarshal(data, &vd); err == nil {
		v.Value = vd.Value
		v.Default = vd.Default
		return nil
	}
	return fmt.Errorf("unmarshal Var failed")
}
