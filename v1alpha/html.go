package v1alpha

import (
	"bytes"
	"context"
	"fmt"

	"github.com/wenooij/nuggit/runtime"
	"golang.org/x/net/html"
)

// HTML parses an HTML document.
type HTML struct {
	Sink  *Sink  `json:"sink,omitempty"`
	Bytes []byte `json:"bytes,omitempty"`
}

func (x *HTML) Bind(e runtime.Edge) error {
	switch e.SrcField {
	case "sink":
		x.Sink = e.Result.(*Sink)
	case "bytes":
		switch e.Result.(type) {
		case string:
			x.Bytes = []byte(e.Result.(string))
		default:
			x.Bytes = e.Result.([]byte)
		}
	case "":
		*x = *e.Result.(*HTML)
	default:
		return fmt.Errorf("not found: %q", e.SrcField)
	}
	return nil
}

func (x *HTML) Run(ctx context.Context) (any, error) {
	if x.Sink != nil && x.Bytes != nil {
		return nil, fmt.Errorf("cannot set both Sink and Bytes")
	}
	data := x.Bytes
	if x.Sink != nil {
		result, err := x.Sink.Run(ctx)
		if err != nil {
			return nil, err
		}
		data = result.([]byte)
	}
	node, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return node, nil
}
