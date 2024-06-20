package runtime

import (
	"context"
	"runtime"

	"github.com/wenooij/nuggit"
)

// MaxWorkers defines the maximum Node-level concurrency in the StageRunner.
var MaxWorkers = runtime.GOMAXPROCS(0)

type Validator interface {
	Validate() error
}

type Runner interface {
	Run(ctx context.Context) (any, error)
}

type OpFactory interface {
	New(nuggit.Node) (any, error)
}
