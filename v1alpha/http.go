package v1alpha

import (
	"context"
	"fmt"
	"net/http"

	"github.com/wenooij/nuggit/keys"
	"github.com/wenooij/nuggit/runtime"
)

func (x *HTTP) Bind(e runtime.Edge) error {
	switch head, tail := keys.Cut(e.SrcField); head {
	case "source":
		if x.Source == nil {
			x.Source = new(Source)
		}
		return x.Source.Bind(e.CloneWithSrcField(tail))
	case "":
		return nil
	default:
		return fmt.Errorf("Bind: unsupported SrcField: %q", e.SrcField)
	}
}

func (x *HTTP) Run(ctx context.Context) (any, error) {
	u, err := x.Source.URL()
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(&http.Request{
		Method: x.Method,
		URL:    u,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}
