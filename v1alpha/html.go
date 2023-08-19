package v1alpha

import (
	"bytes"
	"context"
	"fmt"

	"golang.org/x/net/html"
)

func (x *HTML) Bind(edges []Edge) error {
	if x.Sink == nil {
		return fmt.Errorf("HTML requires a Sink")
	}
	// TODO(wes): Implement alternative to Sink.
	return nil
}

func (r *HTML) Run(ctx context.Context) (any, error) {
	// TODO(wes): Implement alternative to Bytes.
	node, err := html.Parse(bytes.NewReader(r.Sink.Bytes))
	if err != nil {
		return nil, err
	}
	return node, nil
}
