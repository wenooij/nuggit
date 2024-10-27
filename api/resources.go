package api

import (
	"encoding/json"
	"fmt"
	"hash"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/status"
	"gopkg.in/yaml.v3"
)

const resourcesBaseURI = "/api/resources"

type Resource struct {
	APIVersion APIVersion        `json:"api_version,omitempty"`
	Kind       Kind              `json:"kind,omitempty"`
	Metadata   *ResourceMetadata `json:"metadata,omitempty"`
	Spec       any               `json:"spec,omitempty"`
}

func NewResourceSpec(kind Kind) (any, error) {
	switch kind {
	case KindPipe:
		return new(nuggit.Pipe), nil
	case KindView:
		return new(View), nil
	default:
		return nil, fmt.Errorf("unsupported resource kind (%q): %w", kind, status.ErrInvalidArgument)
	}
}

func (r *Resource) GetAPIVersion() APIVersion {
	if r == nil {
		return ""
	}
	return r.APIVersion
}

func (r *Resource) GetKind() Kind {
	if r == nil {
		return ""
	}
	return r.Kind
}

func (r *Resource) GetMetadata() *ResourceMetadata {
	if r == nil {
		return nil
	}
	return r.Metadata
}

func (r *Resource) GetName() string   { return r.GetMetadata().GetName() }
func (r *Resource) GetDigest() string { return r.GetMetadata().GetDigest() }

func (r *Resource) GetSpec() any {
	if r == nil {
		return nil
	}
	return r.Spec
}

func (r *Resource) GetPipe() *nuggit.Pipe {
	if r == nil {
		return nil
	}
	pipe, ok := r.Spec.(*nuggit.Pipe)
	if !ok || pipe == nil {
		return nil
	}
	return pipe
}

func (r *Resource) GetView() *View {
	if r == nil {
		return nil
	}
	c, ok := r.Spec.(*View)
	if !ok {
		return nil
	}
	return c
}

func (r *Resource) ReplaceSpec(spec any) bool {
	if r == nil {
		return false
	}
	r.Spec = spec
	return true
}

func (r *Resource) SetName(name string)     { r.GetMetadata().SetName(name) }
func (r *Resource) SetDigest(digest string) { r.GetMetadata().SetDigest(digest) }

func (r *Resource) UnmarshalJSON(data []byte) error {
	var temp struct {
		APIVersion APIVersion        `json:"api_version,omitempty"`
		Kind       Kind              `json:"kind,omitempty"`
		Metadata   *ResourceMetadata `json:"metadata,omitempty"`
		Spec       json.RawMessage   `json:"spec,omitempty"`
	}
	if temp.Spec == nil {
		temp.Spec = []byte("null")
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("failed to unmarshal resource: %w", err)
	}
	spec, err := NewResourceSpec(temp.Kind)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(temp.Spec, spec); err != nil {
		return fmt.Errorf("failed to unmarshal spec (%q): %w", temp.Kind, err)
	}
	r.APIVersion = temp.APIVersion
	r.Kind = temp.Kind
	r.Metadata = temp.Metadata
	r.Spec = spec
	return nil
}

func (r *Resource) UnmarshalYAML(value *yaml.Node) error {
	var temp struct {
		APIVersion APIVersion        `yaml:"api_version,omitempty"`
		Kind       Kind              `yaml:"kind,omitempty"`
		Metadata   *ResourceMetadata `yaml:"metadata,omitempty"`
		Spec       yaml.Node         `yaml:"spec,omitempty"`
	}
	if err := value.Decode(&temp); err != nil {
		return fmt.Errorf("failed to unmarshal resource: %w", err)
	}
	spec, err := NewResourceSpec(temp.Kind)
	if err != nil {
		return err
	}
	if err := temp.Spec.Decode(spec); err != nil {
		return fmt.Errorf("failed to decode spec (%q): %w", temp.Kind, err)
	}
	r.APIVersion = temp.APIVersion
	r.Kind = temp.Kind
	r.Metadata = temp.Metadata
	r.Spec = spec
	return nil
}

func (r *Resource) writeDigest(h hash.Hash) error {
	return json.NewEncoder(h).Encode(r.GetSpec())
}

type Kind = string

const (
	KindPipe = "pipe"
	KindView = "view"
)

type APIVersion = string

const (
	V1 APIVersion = "v1"
)

type ResourceMetadata struct {
	Name        string   `json:"name,omitempty"`
	Digest      string   `json:"digest,omitempty"`
	Version     string   `json:"version,omitempty"`
	Description string   `json:"description,omitempty"`
	Labels      []string `json:"labels,omitempty"`
}

func (m *ResourceMetadata) GetName() string {
	if m == nil {
		return ""
	}
	return m.Name
}

func (m *ResourceMetadata) GetDigest() string {
	if m == nil {
		return ""
	}
	return m.Digest
}

func (m *ResourceMetadata) GetVersion() string {
	if m == nil {
		return ""
	}
	return m.Version
}

func (m *ResourceMetadata) GetDescription() string {
	if m == nil {
		return ""
	}
	return m.Description
}

func (m *ResourceMetadata) GetLabels() []string {
	if m == nil {
		return nil
	}
	return m.Labels
}

func (m *ResourceMetadata) SetName(name string) {
	if m != nil {
		m.Name = name
	}
}

func (m *ResourceMetadata) SetDigest(digest string) {
	if m != nil {
		m.Digest = digest
	}
}

type ResourcesAPI struct {
}
