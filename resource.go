package nuggit

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

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

// ResourceEncoder encodes a Resource to
type ResourceEncoder struct {
	Gzip bool
}

// Encode encodes the Resource with the encoding specified in Metadata.
//
// See EncodeWithSums to additionally compute checksums.
func (e ResourceEncoder) Encode(r *Resource) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	br := r.BaseResource
	br.Metadata = new(Metadata)
	if r.Metadata != nil {
		*br.Metadata = *r.Metadata
	}
	sort.Strings(br.Metadata.Labels)
	if e.Gzip {
		br.Metadata.Encoding = "gzip"
	}
	base, err := json.Marshal(br)
	if err != nil {
		return nil, err
	}
	buf.Write(base[1 : len(base)-1]) // Hack to remove '{' and '}'.
	spec, err := json.Marshal(r.Spec)
	if err != nil {
		return nil, err
	}
	if len(base) > 2 {
		buf.WriteByte(',')
	}
	buf.WriteString(`"spec":`)
	if e.Gzip {
		var b bytes.Buffer
		zw := gzip.NewWriter(&b)
		zw.Write(spec)
		if err := zw.Close(); err != nil {
			return nil, err
		}
		buf.WriteString(strconv.Quote(b.String()))
	} else {
		buf.Write(spec)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// MarshalJSON marshals the Resource as a JSON object.
func (r *Resource) MarshalJSON() ([]byte, error) {
	return ResourceEncoder{}.Encode(r)
}

// ResourceDecoder decodes a Resource from encoded JSON data.
type ResourceDecoder struct {
	// Sums defines checksums to use on the encoded byte data of a Resource.
	// Only those checksums that are nonempty are validated.
	Sums *Sums
}

// Decode decodes a Resource and validates the checksums, if provided.
func (d ResourceDecoder) Decode(data []byte) (*Resource, error) {
	var rr RawResource
	if err := json.Unmarshal(data, &rr); err != nil {
		return nil, err
	}
	if err := d.Sums.TestBytes(data); err != nil {
		return nil, fmt.Errorf("failed integrity check: %v", err)
	}
	// TODO(wes): Handle gzip decoding.
	var spec any
	switch rr.Kind {
	case "Graph":
		var g Graph
		if err := json.Unmarshal([]byte(rr.Spec), &g); err != nil {
			return nil, err
		}
		spec = &g
	default: // Unknown Kind, simply interpret Spec as default JSON.
		if err := json.Unmarshal([]byte(rr.Spec), &spec); err != nil {
			return nil, err
		}
	}
	var r Resource
	r.BaseResource = rr.BaseResource
	r.Spec = spec
	return &r, nil
}

// UnmarshalJSON implements JSON unmarshaling for Resources given arbitrary Kind and encoding.
// UnmarshalJSON returns an error when the Resource cannot be interpreted as a JSON object at least
// (e.g. map[string]any) or fails to unmarshal as the specified Kind.
func (r *Resource) UnmarshalJSON(data []byte) error {
	v, err := ResourceDecoder{}.Decode(data)
	if err != nil {
		return err
	}
	*r = *v
	return nil
}
