package nuggit

import "encoding/json"

// Metadata defines additional data to attach to an API Resource.
//
// See Resource.
type Metadata struct {
	// Name is a human name for a Resource.
	Name string `json:"name,omitempty"`
	// Description is human text describing a Resource.
	Description string `json:"description,omitempty"`
	// Version is an opaque version ID for a Resource.
	//
	// Examples:
	//
	//	* v1
	//	* v1.1.0-release
	//
	Version string `json:"version,omitempty"`
	// Encoding describes the encoding of a Resource Spec.
	// For instance, if the Spec is a GZip encoded JSON blob
	// "gzip" can be used.
	// An empty string typically means JSON encoded.
	//
	// Examples:
	//
	//	* gzip
	//
	Encoding string `json:"encoding,omitempty"`
	// Labels are opaque strings to describe a Resource.
	// These are usually simple string tags.
	Labels []string `json:"labels,omitempty"`
}

// BaseResource applies common metadata to a Resource or RawResource.
//
// See Resource.
// See RawResource.
type BaseResource struct {
	// APIVersion specifies the version of the Nuggit API to use when evaluating the Spec.
	//
	// Examples:
	//
	//	* v1alpha
	//	* v1
	//
	APIVersion string `json:"api_version,omitempty"`
	// Kind describes the type of the attached Spec.
	//
	// Examples:
	//
	//	* Graph
	//
	Kind string `json:"kind,omitempty"`
	// Metadata describes additional data to associate with the Spec.
	Metadata *Metadata `json:"metadata,omitempty"`
}

// RawResource is used to unmarshal Resource in two phases.
// The first phase unmarshals the envelope and the second the Spec contents.
// The purpose is to support arbitrary Kind, encoding types, and checksums.
//
// See Resource.UnmarshalJSON.
// See Metadata.Encoding.
// See Sums.
type RawResource struct {
	BaseResource `json:",omitempty"`
	// Spec is arbitrary data that this Resource describes.
	Spec json.RawMessage `json:"spec,omitempty"`
}

// Resource is an envelope which adds metadata and integrity checks to an API Spec.
//
// See Metadata.
type Resource struct {
	BaseResource `json:",omitempty"`
	// Spec is arbitrary data that this Resource describes.
	Spec any `json:"spec,omitempty"`
}
