package v1alpha

import (
	"bytes"
	"context"
	"fmt"

	"github.com/wenooij/nuggit/runtime"
	"golang.org/x/net/html"
)

func (x *HTML) Bind(e runtime.Edge) error {
	// TODO(wes): Implement Sink binding.
	return nil
}

func (x *HTML) Run(ctx context.Context) (any, error) {
	if x.Sink == nil {
		return nil, fmt.Errorf("HTML requires a Sink")
	}
	// TODO(wes): Implement alternative to Bytes.
	node, err := html.Parse(bytes.NewReader(x.Sink.Bytes))
	if err != nil {
		return nil, err
	}
	return node, nil
}
