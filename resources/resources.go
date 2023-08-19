package resources

import (
	"encoding/json"
	"os"

	"github.com/wenooij/nuggit"
)

// FromFile decodes a Resource from the given local file.
// Checksums are ignored.
func FromFile(filename string) (*nuggit.Resource, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return nuggit.ResourceDecoder{}.Decode(data)
}

// FromBytes decodes a Resource from the given byte slice.
// Checksums are ignored.
func FromBytes(data []byte) (*nuggit.Resource, error) {
	return nuggit.ResourceDecoder{}.Decode(data)
}

// ReadGraph reads a graph from a JSON file containing either a Resource
// or Graph definition.
func ReadGraph(filename string) (*nuggit.Graph, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var r nuggit.Resource
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	switch r.Kind {
	case "Graph":
		return r.Spec.(*nuggit.Graph), nil
	default: // Try Graph.UnmarshalJSON.
		var g nuggit.Graph
		if err := json.Unmarshal(data, &g); err != nil {
			return nil, err
		}
		return &g, nil
	}
}
