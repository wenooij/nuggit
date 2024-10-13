package nuggit

import (
	"encoding/json"
	"reflect"
)

type Pipeline struct {
	RunCondition *RunCondition `json:"conditions,omitempty"`
	Sequence     []string      `json:"sequence,omitempty"`
}

func (p Pipeline) Root() (string, bool) {
	if len(p.Sequence) == 0 {
		return "", false
	}
	return p.Sequence[0], true
}

type RunCondition struct {
	AlwaysEnabled bool   `json:"always_enabled,omitempty"`
	Host          string `json:"host,omitempty"`
	URLPattern    string `json:"url_pattern,omitempty"`
}

type DataSpecifier struct {
	Collection string `json:"collection,omitempty"`
	Point      string `json:"point,omitempty"`
}

type RawNode struct {
	Action string          `json:"action,omitempty"`
	Spec   json.RawMessage `json:"spec,omitempty"`
}

type Node[T any] struct {
	Action string `json:"action,omitempty"`
	Spec   T      `json:"spec,omitempty"`
}

func (n Node[T]) Raw() (RawNode, error) {
	raw := RawNode{Action: n.Action}
	if reflect.ValueOf(n.Spec).IsZero() {
		return raw, nil
	}
	spec, err := json.Marshal(n.Spec)
	if err != nil {
		return RawNode{}, err
	}
	raw.Spec = spec
	return raw, nil
}

type Type string

const (
	TypeUnspecified Type = "" // Same as TypeBytes.
	TypeBytes
	TypeString
	TypeBool
	TypeInt64
	TypeUint64
	TypeFloat64
	TypeBigInt
	TypeBigFloat
	TypeDOMElement
	TypeDOMTree
)
