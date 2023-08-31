package v1alpha

import (
	"context"
	"fmt"
	"io"
)

type Sink struct {
	Op         SinkOp `json:"op,omitempty"`
	BufferSize int    `json:"buffer_size,omitempty"`
	Offset     int    `json:"offset,omitempty"`
	Bytes      []byte `json:"bytes,omitempty"`
	String     string `json:"string,omitempty"`
	HTTP       *HTTP  `json:"http,omitempty"`
	File       *File  `json:"file,omitempty"`
}

func (x *Sink) Run(ctx context.Context) (v any, err error) {
	switch x.Op {
	case SinkUndefined, SinkBytes:
		return x.Bytes, nil
	case SinkString:
		return []byte(x.String), nil
	case SinkHTTP:
		resp, err := x.HTTP.Response()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(resp.Body)
		defer func() {
			if err1 := resp.Body.Close(); err == nil && err1 != nil {
				err = err1
			}
		}()
		if err != nil {
			return nil, err
		}
		return data, nil
	case SinkFile:
		data, err := x.File.ReadFile()
		if err != nil {
			return nil, err
		}
		return data, nil
	default:
		return nil, fmt.Errorf("unknown Sink Op: %q", x.Op)
	}
}

type SinkOp string

const (
	SinkUndefined SinkOp = ""
	SinkBytes     SinkOp = "bytes"
	SinkString    SinkOp = "string"
	SinkHTTP      SinkOp = "http"
	SinkFile      SinkOp = "file"
)
