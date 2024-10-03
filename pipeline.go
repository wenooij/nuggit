package nuggit

import "encoding/json"

type Pipeline struct {
	Name       string     `json:"name,omitempty"`
	Conditions Conditions `json:"conditions,omitempty"`
	Export     Export     `json:"export,omitempty"`
	Ops        []RawOp    `json:"ops,omitempty"`
}

type Conditions struct {
	Host    string   `json:"host,omitempty"`
	Hosts   []string `json:"hosts,omitempty"`
	Pattern string   `json:"pattern,omitempty"`
}

type Export struct {
	Nullable        bool          `json:"nullable,omitempty"`
	IncludeMetadata bool          `json:"include_metadata,omitempty"`
	ID              DataSpecifier `json:"id,omitempty"`
}

type DataSpecifier struct {
	Collection string `json:"collection,omitempty"`
	Name       string `json:"name,omitempty"`
}

type RawOp struct {
	Action string          `json:"action,omitempty"`
	Spec   json.RawMessage `json:"spec,omitempty"`
}

type Op[T any] struct {
	Action string `json:"action,omitempty"`
	Spec   T      `json:"spec,omitempty"`
}
