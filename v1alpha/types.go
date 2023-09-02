package v1alpha

import "github.com/wenooij/nuggit"

type Field struct {
	Key  nuggit.FieldKey
	Type *Type
}

type Type struct {
	Key    string      `json:"key,omitempty"`
	Type   nuggit.Type `json:"type,omitempty"`
	Fields []Field     `json:"fields,omitempty"`
}
