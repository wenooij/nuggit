package v1alpha

import (
	"context"
	"fmt"

	"github.com/wenooij/nuggit"
)

type Median struct {
	Type        nuggit.Type `json:"type,omitempty"`
	CompareKey  string      `json:"compare_key"`
	Values      []any       `json:"any_values,omitempty"`
	TupleIndex  int         `json:"tuple_index,omitempty"`
	TupleValues [][]any     `json:"tuple_values,omitempty"`
}

func (x *Median) Run(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}
