package v1alpha

import "github.com/wenooij/nuggit"

type Field struct {
	Key  string
	Type *Type
}

type Type struct {
	Key    string      `json:"key,omitempty"`
	Type   nuggit.Type `json:"type,omitempty"`
	Fields []Field     `json:"fields,omitempty"`
}
