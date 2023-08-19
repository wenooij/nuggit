package v1alpha

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func (x *Sink) Bind(edges []Edge) error {
	for _, e := range edges {
		switch res := e.Result.(type) {
		case *http.Response:
			data, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}
			res.Body.Close()
			x.Bytes = data
			// TODO(wes): Handle other sinking ops.
		default:
			return fmt.Errorf("Sink: unexpected type in input: %T", res)
		}
	}

	return nil
}

func (x *Sink) Run(ctx context.Context) (any, error) {
	var resp *http.Response

	if resp == nil {
		return nil, fmt.Errorf("expected http.Response as input")
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}
