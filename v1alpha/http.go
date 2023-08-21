package v1alpha

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/wenooij/nuggit/runtime"
)

func (x *HTTP) Bind(edges []runtime.Edge) error {
	for _, e := range edges {
		switch e.SrcField {
		case "source.host":
			if err := x.BindHost(e); err != nil {
				return err
			}
		case "source.path":
			if err := x.BindPath(e); err != nil {
				return err
			}
		case "source":
			if err := x.BindSource(e); err != nil {
				return err
			}
		case "":
			// TODO(wes): Handle issues of merging.
			if err := json.Unmarshal(e.Data, &x); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Bind: unsupported SrcField: %q", e.SrcField)
		}
	}
	return nil
}

func (x *HTTP) BindHost(e runtime.Edge) error {
	if x.Source == nil {
		x.Source = new(Source)
	}
	switch res := e.Result.(type) {
	case string:
		x.Source.Host = res
		return nil
	default:
		return fmt.Errorf("BindHost: unsupported type: %T", res)
	}
}

func (x *HTTP) BindPath(e runtime.Edge) error {
	if x.Source == nil {
		x.Source = new(Source)
	}
	switch res := e.Result.(type) {
	case string:
		x.Source.Path = res
		return nil
	default:
		return fmt.Errorf("BindPath: unsupported type: %T", res)
	}
}

func (x *HTTP) BindSource(e runtime.Edge) error {
	if x.Source == nil {
		x.Source = new(Source)
	}
	switch res := e.Result.(type) {
	case string:
		u, err := url.Parse(res)
		if err != nil {
			return err
		}
		*x.Source = Source{
			Host:  u.Host,
			Path:  u.Path,
			Query: u.RawQuery,
		}
		return nil
	case *url.URL:
		*x.Source = Source{
			Host:  res.Host,
			Path:  res.Path,
			Query: res.RawQuery,
		}
		return nil
	default:
		return fmt.Errorf("BindHost: unsupported type: %T", res)
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
