package v1alpha

import (
	"context"
	"fmt"

	"github.com/andybalholm/cascadia"
	"github.com/wenooij/nuggit/runtime"
	"golang.org/x/net/html"
)

// Selector implements CSS selectors.
type Selector struct {
	All      bool       `json:"all,omitempty"`
	Selector string     `json:"selector,omitempty"`
	Node     *html.Node `json:"-"`
	Bytes    []byte     `json:"bytes,omitempty"`
}

func (x *Selector) Bind(e runtime.Edge) error {
	switch e.SrcField {
	case "all":
		x.All = e.Result.(bool)
	case "selector":
		x.Selector = e.Result.(string)
	case "node":
		x.Node = e.Result.(*html.Node)
	case "bytes":
		switch e.Result.(type) {
		case string:
			x.Bytes = []byte(e.Result.(string))
		default:
			x.Bytes = e.Result.([]byte)
		}
	case "":
		*x = *e.Result.(*Selector)
	default:
		return fmt.Errorf("not found: %q", e.SrcField)
	}
	return nil
}

func (x *Selector) Run(ctx context.Context) (any, error) {
	if x.Node == nil {
		return nil, fmt.Errorf("missing Node")
	}
	sel, err := cascadia.Compile(x.Selector)
	if err != nil {
		return nil, err
	}
	if x.All {
		return cascadia.QueryAll(x.Node, sel), nil
	}
	return cascadia.Query(x.Node, sel), nil
}
