package v1alpha

import (
	"context"
	"fmt"
	"net/url"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/keys"
	"github.com/wenooij/nuggit/runtime"
)

// Source defines a Web source with a given host and path.
//
// The path elements are dynamic to support variables and step outputs.
// The Host is a static variable not changeable through inputs.
type Source struct {
	Scheme string `json:"scheme,omitempty"`
	Host   string `json:"host,omitempty"`
	Path   string `json:"path,omitempty"`
	Query  string `json:"query,omitempty"`
}

func (x *Source) URL() (*url.URL, error) {
	scheme := x.Scheme
	if scheme == "" {
		scheme = "http"
	}
	res, err := url.JoinPath(
		fmt.Sprint(scheme, "://", x.Host),
		x.Path,
	)
	if err != nil {
		return nil, err
	}
	// TODO(wes): Join Query.
	return url.Parse(res)
}

func (x *Source) Bind(e runtime.Edge) error {
	switch head, tail := keys.Cut(e.SrcField); head {
	case "scheme":
		if tail != "" {
			return fmt.Errorf("src_field not found: %q", e.SrcField)
		}
		x.Scheme = e.Result.(string)
		return nil
	case "host":
		if tail != "" {
			return fmt.Errorf("src_field not found: %q", e.SrcField)
		}
		x.Host = e.Result.(string)
		return nil
	case "path":
		if tail != "" {
			return fmt.Errorf("src_field not found: %q", e.SrcField)
		}
		switch e.Glom {
		case nuggit.GlomAppend:
			x.Path = fmt.Sprint(x.Path, e.Result.(string))
		default:
			x.Path = e.Result.(string)
		}
		return nil
	case "query":
		if tail != "" {
			return fmt.Errorf("src_field not found: %q", e.SrcField)
		}
		x.Query = e.Result.(string)
		return nil
	case "":
		*x = *e.Result.(*Source)
		return nil
	default:
		return fmt.Errorf("src_field not found: %q", head)
	}
}

func (x *Source) Run(ctx context.Context) (any, error) {
	return x, nil
}
