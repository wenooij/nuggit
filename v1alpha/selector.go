package v1alpha

import (
	"bytes"
	"context"
	"fmt"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

// Selector implements CSS selectors.
type Selector struct {
	All      bool   `json:"all,omitempty"`
	Selector string `json:"selector,omitempty"`
	Sink     *Sink  `json:"sink,omitempty"`
}

func (x *Selector) Run(ctx context.Context) (any, error) {
	if x.Sink == nil {
		return nil, fmt.Errorf("Sink is required")
	}
	data, err := x.Sink.Run(ctx)
	if err != nil {
		return nil, err
	}
	// TODO(wes): Handle unchecked cast to bytes.
	node, err := html.Parse(bytes.NewReader(data.([]byte)))
	if err != nil {
		return nil, err
	}
	sel, err := cascadia.ParseWithPseudoElement(x.Selector)
	if err != nil {
		return nil, err
	}
	var res [][]byte
	var nodes []*html.Node
	if x.All {
		nodes = cascadia.QueryAll(node, sel)
	} else {
		if n := cascadia.Query(node, sel); n != nil {
			nodes = []*html.Node{n}
		}
	}
	for _, n := range nodes {
		var buf bytes.Buffer
		html.Render(&buf, n)
		res = append(res, buf.Bytes())
	}
	return res, nil
}
