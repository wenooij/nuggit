package v1alpha

import (
	"bytes"
	"context"
	"fmt"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// HTML parses an HTML document.
type HTML struct {
	Sink   *Sink  `json:"sink,omitempty"`
	String string `json:"string,omitempty"`
	Bytes  []byte `json:"bytes,omitempty"`
}

func (x *HTML) Run(ctx context.Context) (any, error) {
	if x.Sink != nil && x.Bytes != nil {
		return nil, fmt.Errorf("cannot set both Sink and Bytes")
	}
	if x.Sink != nil && x.String != "" {
		return nil, fmt.Errorf("cannot set both Sink and String")
	}
	if x.Bytes != nil && x.String != "" {
		return nil, fmt.Errorf("cannot set both Bytes and String")
	}
	data := x.Bytes
	if x.Sink != nil {
		result, err := x.Sink.Run(ctx)
		if err != nil {
			return nil, err
		}
		data = result.([]byte)
	}
	if data == nil {
		data = []byte(x.String)
	}
	node, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	// TODO(wes): Temporary situation while we find out how to marshal HTML nodes
	//            and whether HTML Node has a purpose still.
	return struct {
		Type      html.NodeType    `json:"type,omitempty"`
		DataAtom  atom.Atom        `json:"atom,omitempty"`
		Data      string           `json:"data,omitempty"`
		Namespace string           `json:"namespace,omitempty"`
		Attr      []html.Attribute `json:"attr,omitempty"`
	}{
		Type:      node.Type,
		DataAtom:  node.DataAtom,
		Data:      node.Data,
		Namespace: node.Namespace,
		Attr:      node.Attr,
	}, nil
}
