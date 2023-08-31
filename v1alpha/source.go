package v1alpha

import (
	"fmt"
	"net/url"
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
