package v1alpha

import (
	"context"
	"fmt"

	"github.com/wenooij/nuggit"
)

type Runner struct{}

func (Runner) Run(ctx context.Context, n nuggit.Node) (any, error) {
	opFn, ok := runnerMap[n.Op]
	if !ok {
		return nil, fmt.Errorf("op is not defined for Node %v: %q", n.Key, n.Op)
	}
	return opFn(ctx, n.Data)
}

var runnerMap = map[nuggit.Op]func(ctx context.Context, data any) (any, error){
	nuggit.OpAssert:   func(ctx context.Context, data any) (any, error) { return nil, fmt.Errorf("not implemented") },
	nuggit.OpCache:    func(ctx context.Context, data any) (any, error) { return nil, fmt.Errorf("not implemented") },
	nuggit.OpChromedp: func(ctx context.Context, data any) (any, error) { return nil, fmt.Errorf("not implemented") },
	nuggit.OpCond:     func(ctx context.Context, data any) (any, error) { return nil, fmt.Errorf("not implemented") },
	nuggit.OpFile:     func(ctx context.Context, data any) (any, error) { return nil, fmt.Errorf("not implemented") },
	nuggit.OpFind:     func(ctx context.Context, data any) (any, error) { return nil, fmt.Errorf("not implemented") },
	nuggit.OpGroupBy:  func(ctx context.Context, data any) (any, error) { return nil, fmt.Errorf("not implemented") },
	nuggit.OpHTML:     func(ctx context.Context, data any) (any, error) { return nil, fmt.Errorf("not implemented") },
	nuggit.OpHTTP:     func(ctx context.Context, data any) (any, error) { return nil, fmt.Errorf("not implemented") },
	nuggit.OpPoint:    func(ctx context.Context, data any) (any, error) { return nil, fmt.Errorf("not implemented") },
	nuggit.OpRegex:    func(ctx context.Context, data any) (any, error) { return nil, fmt.Errorf("not implemented") },
	nuggit.OpReplace:  func(ctx context.Context, data any) (any, error) { return nil, fmt.Errorf("not implemented") },
	nuggit.OpSelect:   func(ctx context.Context, data any) (any, error) { return nil, fmt.Errorf("not implemented") },
	nuggit.OpSink:     func(ctx context.Context, data any) (any, error) { return nil, fmt.Errorf("not implemented") },
	nuggit.OpSource:   func(ctx context.Context, data any) (any, error) { return nil, fmt.Errorf("not implemented") },
	nuggit.OpSpan:     func(ctx context.Context, data any) (any, error) { return nil, fmt.Errorf("not implemented") },
	nuggit.OpTable:    func(ctx context.Context, data any) (any, error) { return nil, fmt.Errorf("not implemented") },
	nuggit.OpTime:     func(ctx context.Context, data any) (any, error) { return nil, fmt.Errorf("not implemented") },
	nuggit.OpVar:      func(ctx context.Context, data any) (any, error) { return nil, fmt.Errorf("not implemented") },
}
