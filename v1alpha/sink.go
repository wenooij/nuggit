package v1alpha

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/wenooij/nuggit/keys"
	"github.com/wenooij/nuggit/runtime"
)

type Sink struct {
	BufferSize int       `json:"buffer_size,omitempty"`
	Offset     int       `json:"offset,omitempty"`
	Bytes      []byte    `json:"bytes,omitempty"`
	Reader     io.Reader `json:"-"`
	Sink       *Sink     `json:"sink,omitempty"`
}

func (x *Sink) Bind(e runtime.Edge) error {
	switch head, tail := keys.Cut(e.SrcField); head {
	case "buffer_size", "offset", "bytes", "sink":
		if tail != "" {
			return fmt.Errorf("unexpected field: %v", e.SrcField)
		}
		panic("not implemented")
	case "reader":
		if tail != "" {
			return fmt.Errorf("unexpected field: %v", e.SrcField)
		}
		x.Reader = e.Result.(io.Reader)
		return nil
	case "":
		// Infer from result type.
		switch res := e.Result.(type) {
		case *http.Response:
			// TODO(wes): Handle SrcField.
			x.Reader = res.Body
			return nil
		default:
			return fmt.Errorf("unexpected type in input: %T", res)
		}
	default:
		return fmt.Errorf("unknown field: %v", head)
	}
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
