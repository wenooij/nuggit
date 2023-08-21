package runtime

import (
	"context"

	"github.com/wenooij/nuggit"
)

type Edge struct {
	nuggit.Edge
	Result any
}

type Binder interface {
	Bind(edges []Edge) error
}

type Runner interface {
	Run(ctx context.Context) (any, error)
}

type RunnerFactory interface {
	NewRunner(nuggit.Node) (Runner, error)
}
