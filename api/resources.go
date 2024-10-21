package api

import (
	"encoding/json"
)

const resourcesBaseURI = "/api/resources"

type Resource struct {
	ApiVersion string            `json:"api_version,omitempty"`
	Kind       string            `json:"kind,omitempty"`
	Metadata   *ResourceMetadata `json:"metadata,omitempty"`
	Spec       json.RawMessage   `json:"spec,omitempty"`
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
