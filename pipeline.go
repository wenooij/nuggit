package nuggit

import "encoding/json"

type Pipeline struct {
	Name         string       `json:"name,omitempty"`
	RunCondition RunCondition `json:"conditions,omitempty"`
	Export       Export       `json:"export,omitempty"`
	Ops          []RawOp      `json:"ops,omitempty"`
}

type RunCondition struct {
	Host       string   `json:"host,omitempty"`
	Hosts      []string `json:"hosts,omitempty"`
	URLPattern string   `json:"url_pattern,omitempty"`
}

type Export struct {
	Nullable        bool          `json:"nullable,omitempty"`
	IncludeMetadata bool          `json:"include_metadata,omitempty"`
	Type            Type          `json:"type,omitempty"`
	ID              DataSpecifier `json:"id,omitempty"`
}

type DataSpecifier struct {
	Collection string `json:"collection,omitempty"`
	Point      string `json:"point,omitempty"`
}

type RawOp struct {
	Action string          `json:"action,omitempty"`
	Spec   json.RawMessage `json:"spec,omitempty"`
}

type Op[T any] struct {
	Action string `json:"action,omitempty"`
	Spec   T      `json:"spec,omitempty"`
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
