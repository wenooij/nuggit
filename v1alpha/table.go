package v1alpha

import "github.com/wenooij/nuggit"

type Col struct {
	Name     string      `json:"name,omitempty"`
	Type     nuggit.Type `json:"type,omitempty"`
	Nullable bool        `json:"nullable,omitempty"`
}

type Table struct {
	Cols []Col    `json:"cols,omitempty"`
	Key  []string `json:"key,omitempty"`
}

func (x *Table) Unmarshal(data any) error { return nil }
