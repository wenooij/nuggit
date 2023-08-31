package v1alpha

import (
	"net/http"
)

type HTTP struct {
	Source *Source           `json:"source,omitempty"`
	Method string            `json:"method,omitempty"`
	Header map[string]string `json:"header,omitempty"`
}

func (x *HTTP) GetHeader() http.Header {
	header := make(http.Header, len(x.Header))
	for k, v := range x.Header {
		header[k] = []string{v}
	}
	return header
}

func (x *HTTP) Request() (*http.Request, error) {
	u, err := x.Source.URL()
	if err != nil {
		return nil, err
	}
	return &http.Request{
		Method: x.Method,
		Header: x.GetHeader(),
		URL:    u,
	}, nil
}

func (x *HTTP) Response() (*http.Response, error) {
	req, err := x.Request()
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func (x *HTTP) Validate() error {
	return nil
}
