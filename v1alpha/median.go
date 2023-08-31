package v1alpha

import (
	"context"
	"fmt"

	"github.com/wenooij/nuggit"
)

type Median struct {
	Type        nuggit.Type     `json:"type,omitempty"`
	Values      []*Const        `json:"values,omitempty"`
	CompareKey  nuggit.FieldKey `json:"compare_key"`
	AnyValues   []Any           `json:"any_values,omitempty"`
	TupleIndex  int             `json:"tuple_index,omitempty"`
	TupleValues [][]Any         `json:"tuple_values,omitempty"`
}

func (x *Median) Run(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
