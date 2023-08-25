package v1alpha

import (
	"context"
	"fmt"
	"io"

	"github.com/wenooij/nuggit/runtime"
)

type Sink struct {
	BufferSize int       `json:"buffer_size,omitempty"`
	Offset     int       `json:"offset,omitempty"`
	Reader     io.Reader `json:"-"`
}

func (x *Sink) Bind(e runtime.Edge) error {
	switch e.SrcField {
	case "buffer_size":
		x.BufferSize = e.Result.(int)
	case "offset":
		x.Offset = e.Result.(int)
	case "reader":
		x.Reader = e.Result.(io.Reader)
	case "":
		*x = *e.Result.(*Sink)
	default:
		return fmt.Errorf("not found: %q", e.SrcField)
	}
	return nil
}

func (x *Sink) Run(ctx context.Context) (v any, err error) {
	if x.Reader == nil {
		return nil, fmt.Errorf("missing Reader")
	}
	data, err := io.ReadAll(x.Reader)
	defer func() {
		if rd, ok := x.Reader.(io.Closer); ok {
			if err1 := rd.Close(); err == nil && err1 != nil {
				err = err1
			}
		}
	}()
	if err != nil {
		return nil, err
	}
	return data, nil
}
