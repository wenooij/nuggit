package v1alpha

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/wenooij/nuggit/runtime"
)

func (x *Sink) Bind(edges []runtime.Edge) error {
	for _, e := range edges {
		switch res := e.Result.(type) {
		case *http.Response:
			// TODO(wes): Handle SrcField.
			x.Reader = res.Body
		default:
			return fmt.Errorf("Sink: unexpected type in input: %T", res)
		}
	}
	return nil
}

func (x *Sink) Run(ctx context.Context) (any, error) {
	if x.Bytes != nil {
		return x.Bytes, nil
	}
	data, err := io.ReadAll(x.Reader)
	defer func() {
		if rd, ok := x.Reader.(io.Closer); ok {
			rd.Close()
		}
	}()
	if err != nil {
		return nil, err
	}
	return data, nil
}
