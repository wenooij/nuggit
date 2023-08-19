package v1alpha

import (
	"context"
	"net/url"
)

type SourceData struct {
	Host  string `json:"host,omitempty"`
	Path  string `json:"path,omitempty"`
	Query string `json:"query,omitempty"`
}

// Source defines a Web source with a given host and path.
//
// The path elements are dynamic to support variables and step outputs.
// The Host is a static variable not changeable through inputs.
type Source struct {
	SourceData `json:",omitempty"`
}

func (x *Source) URL() (*url.URL, error) {
	return &url.URL{
		Scheme:   "http",
		Host:     x.Host,
		Path:     x.Path,
		RawQuery: x.Query,
	}, nil
}

func (x *Source) Bind(edges []Edge) error {
	// TODO(wes): Implement string bindings.
	return nil
}

func (x *Source) Run(ctx context.Context) (any, error) {
	return x, nil
}
