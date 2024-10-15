package api

import (
	"encoding/json"

	"github.com/wenooij/nuggit/status"
)

type ResourceLite struct {
	*Ref `json:",omitempty"`
}

type ResourceBase struct {
	ApiVersion string          `json:"api_version,omitempty"`
	Kind       string          `json:"kind,omitempty"`
	Metadata   *Metadata       `json:"metadata,omitempty"`
	Spec       json.RawMessage `json:"spec,omitempty"`
}

type Resource struct {
	*ResourceLite `json:",omitempty"`
	*ResourceBase `json:",omitempty"`
}

type Kind = string

const (
	KindUndefined  = ""
	KindNode       = "node"
	KindPipe       = "pipe"
	KindCollection = "collection"
)

type Version = string

const (
	VersionUndefined = "" // Same as v1.
	V1               = "v1"
)

type Metadata struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Labels      []string `json:"labels,omitempty"`
}

type ResourcesAPI struct {
	resources map[string]*Resource
}

func (r *ResourcesAPI) Init(storeType StorageType) {
	*r = ResourcesAPI{
		resources: make(map[string]*Resource),
	}
}

type CreateResourceRequest struct {
	*ResourceBase `json:"resource,omitempty"`
}

type CreateResourceResponse struct {
	*ResourceLite `json:"resource,omitempty"`
}

func (r *ResourcesAPI) CreateResource(*CreateResourceRequest) (*CreateResourceResponse, error) {
	return nil, status.ErrUnimplemented
}
