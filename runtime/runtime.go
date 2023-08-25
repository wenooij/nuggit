package runtime

import (
	"context"

	"github.com/wenooij/nuggit"
)

type Edge struct {
	nuggit.Edge
	Result any
}

// CloneWithSrcField returns a copy of e with the new SrcField.
// Tail is useful for implementing recursive subfield binders.
//
// See keys.Cut.
// See Binder.
//
// Deprecated: Would be nice to not do this since Result can't easily be cloned.
func (e Edge) CloneWithSrcField(srcField nuggit.FieldKey) Edge {
	edge := e.Edge.Clone()
	edge.SrcField = srcField
	return Edge{
		Edge:   edge,
		Result: e.Result,
	}
}

type Binder interface {
	Bind(Edge) error
}

type Runner interface {
	Run(ctx context.Context) (any, error)
}

type RunnerFactory interface {
	NewRunner(nuggit.Node) (Runner, error)
}
