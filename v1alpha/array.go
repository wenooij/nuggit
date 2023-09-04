package v1alpha

import "github.com/wenooij/nuggit"

type Array struct {
	Type nuggit.Type `json:"type,omitempty"`
	Op   ArrayOp     `json:"op,omitempty"`
}

type ArrayOp string

const (
	ArrayUndefined ArrayOp = ""
	ArrayNop       ArrayOp = "nop"
	ArrayAgg       ArrayOp = "agg"
)
