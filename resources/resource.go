package resources

import "encoding/json"

// Resource enables
type Resource struct {
	ApiVersion string          `json:"api_version,omitempty"`
	Kind       string          `json:"kind,omitempty"`
	Metadata   *Metadata       `json:"metadata,omitempty"`
	Spec       json.RawMessage `json:"spec,omitempty"`
}

type Kind = string

const (
	KindUndefined = ""
	KindNode      = "node"
	KindPipeline  = "pipeline"
)

type Version = string

const (
	VersionUndefined = "" // Same as v1.
	Version1         = "v1"
)

type Metadata struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Labels      []string `json:"labels,omitempty"`
}
