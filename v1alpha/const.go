package v1alpha

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wenooij/nuggit"
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

func (x *Const) Bind([]Edge) error { return nil }

func (x *Const) Run(context.Context) (any, error) { return x, nil }

func (x *Const) UnmarshalJSON(data []byte) error {
	t := struct {
		Type  nuggit.Type `json:"type,omitempty"`
		Value any         `json:"value,omitempty"`
	}{}
	if err := json.Unmarshal(data, &t); err == nil {
		x.Type = t.Type
		x.Value = t.Value
		return nil
	}

	for _, e := range []struct {
		nuggit.Type
		Value any
	}{
		{nuggit.TypeBool, false},
		{nuggit.TypeUint8, uint8(0)},
		{nuggit.TypeInt8, uint8(0)},
		{nuggit.TypeUint16, uint16(0)},
		{nuggit.TypeInt16, int16(0)},
		{nuggit.TypeUint32, uint32(0)},
		{nuggit.TypeInt32, int32(0)},
		{nuggit.TypeUint64, uint64(0)},
		{nuggit.TypeInt64, int64(0)},
		{nuggit.TypeFloat64, float64(0)},
		{nuggit.TypeString, ""},
		// TODO(wes): Add Time.
	} {
		if err := json.Unmarshal(data, &e.Value); err == nil {
			fmt.Println("hi", e.Type, e.Value)
			x.Type = e.Type
			x.Value = e.Value
			return nil
		}
	}

	return fmt.Errorf("unrecognized type")
}
