package api

import (
	"encoding/json"
)

type ResourceLite struct {
	*Ref `json:",omitempty"`
}

func NewResourceLite(id string) *ResourceLite {
	return &ResourceLite{newRef("/api/resources/", id)}
}

type ResourceBase struct {
	ApiVersion string            `json:"api_version,omitempty"`
	Kind       string            `json:"kind,omitempty"`
	Metadata   *ResourceMetadata `json:"metadata,omitempty"`
	Spec       json.RawMessage   `json:"spec,omitempty"`
}

type Resource struct {
	*ResourceLite `json:",omitempty"`
	*ResourceBase `json:",omitempty"`
}

type Kind = string

const (
	KindUndefined  = ""
	KindAction     = "action"
	KindPipe       = "pipe"
	KindCollection = "collection"
)

type Version = string

const (
	VersionUndefined = "" // Same as v1.
	V1               = "v1"
)

type ResourceMetadata struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Labels      []string `json:"labels,omitempty"`
}
