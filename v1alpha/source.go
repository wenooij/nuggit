package v1alpha

import (
	"context"
	"net/url"

	"github.com/wenooij/nuggit/runtime"
)

// Source defines a Web source with a given host and path.
//
// The path elements are dynamic to support variables and step outputs.
// The Host is a static variable not changeable through inputs.
type Source struct {
	Host  string `json:"host,omitempty"`
	Path  string `json:"path,omitempty"`
	Query string `json:"query,omitempty"`
}

func (x *Source) URL() (*url.URL, error) {
	return &url.URL{
		Scheme:   "http",
		Host:     x.Host,
		Path:     x.Path,
		RawQuery: x.Query,
	}, nil
}

func (x *Source) Bind(edges []runtime.Edge) error {
	// TODO(wes): Implement string bindings.
	return nil
}

func (x *Source) Run(ctx context.Context) (any, error) {
	return x, nil
}
