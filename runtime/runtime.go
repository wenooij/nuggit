package runtime

import (
	"context"
	"encoding/json"

	"github.com/wenooij/nuggit"
)

type Validator interface {
	Validate() error
}

type Runner interface {
	Run(ctx context.Context) (any, error)
}

type NodeFactory interface {
	New(nuggit.Node) (any, error)
}

type VarBinder interface {
	Bind(data json.RawMessage, val any) (json.RawMessage, error)
}
