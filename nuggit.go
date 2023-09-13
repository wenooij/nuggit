// Package nuggit provides a declarative API for information retrieval (IR).
package nuggit

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"

	"golang.org/x/exp/slices"
)

type (
	// Key is an alias for any string Key type.
	Key = string
	// NodeKey is an alias for a string Key type describing a Node.
	NodeKey = Key
	// StageKey is an alias for a string Key type describing a Stage.
	StageKey = Key
	// EdgeKey is an alias for a string Key type describing an Edge.
	EdgeKey = Key
	// FieldKey is an alias for a string Key type describing a field.
	//
	// See Edge.SrcField and Edge.DstField.
	FieldKey = Key
)

// Adjacency describes the edges that are adjacent to a given node at Key.
type Adjacency struct {
	// Key describes the src Node.
	Key NodeKey `json:"key,omitempty"`
	// Edges lists all adjacent edges by EdgeKey.
	Edges []EdgeKey `json:"edges,omitempty"`
}

func (a Adjacency) Clone() Adjacency {
	return Adjacency{
		Key:   a.Key,
		Edges: slices.Clone(a.Edges),
	}
}

// Stage describes a sequence of op Nodes which are executed together.
type Stage struct {
	// Key is a unique id for the Stage.
	Key StageKey `json:"key,omitempty"`
	// Nodes is an ordered list of the Nodes in this stage.
	Nodes []NodeKey `json:"nodes,omitempty"`
}

func (s Stage) Clone() Stage {
	return Stage{
		Key:   s.Key,
		Nodes: slices.Clone(s.Nodes),
	}
}

// Edge describes a directed connection between Nodes Src and Dst and, if specified,
// the fields SrcField and DstField. The Glom operation describes how data flows
// between SrcField and DstField when there are multiple possible semantics.
// It is possible to attach arbitrary JSON data to an edge using the Data field.
// In the context of a Graph, it is possible to have multiple Edges between the same
// NodeKey and FieldKeys.
//
// See GlomOp.
type Edge struct {
	// Key is a unique id for the Edge.
	Key EdgeKey `json:"key,omitempty"`
	// Src describes the starting Node.
	Src NodeKey `json:"src,omitempty"`
	// Dst describes the terminal Node.
	Dst NodeKey `json:"dst,omitempty"`
	// SrcField describes the field within Src, if set.
	//
	// See the doc for specific Ops to see which fields are defined.
	SrcField FieldKey `json:"src_field,omitempty"`
	// DstField describes the field within Dst, if set.
	//
	// See the doc for specific Ops to see which fields are defined.
	DstField FieldKey `json:"dst_field,omitempty"`
	// Data specifies arbitrary JSON data to attach to this Edge.
	//
	// The Op runtime may use Data to change the semantics of the connection.
	Data any `json:"data,omitempty"`
}

// Clone returns a shallow copy of e.
func (e Edge) Clone() Edge {
	return Edge{
		Key:      e.Key,
		Src:      e.Src,
		Dst:      e.Dst,
		SrcField: e.SrcField,
		DstField: e.DstField,
		Data:     e.Data,
	}
}

// Node describes a node in a Graph.
// Data provides additional configuration to the Op.
//
// See Op documentation to see which Ops are defined.
type Node struct {
	Key NodeKey `json:"key,omitempty"`
	// Op specifies
	Op Op `json:"op,omitempty"`
	// Data specifies arbitrary data to attach to this Edge.
	//
	// Data can be used to alter the behavior of the Op.
	Data any `json:"data,omitempty"`
}

// Deprecated: Clone can't work with any Data.
func (n Node) Clone() Node {
	return Node{
		Key:  n.Key,
		Op:   n.Op,
		Data: n.Data,
	}
}

// Graph describes a declarative Program DAG.
//
// See Op specfic documentation to see which ops are available.
type Graph struct {
	// Stage is the name of the stage for this Graph.
	Stage StageKey `json:"stage,omitempty"`
	// Adjacency describes the adjacency list for the Graph.
	// Adjacencies relate NodeKeys to the ordered lists of EdgeKeys sourced there.
	//
	// See Adjacency.
	Adjacency []Adjacency `json:"adjacency,omitempty"`
	// Edges describe the edge list for the Graph.
	//
	// See Edge.
	Edges []Edge `json:"edges,omitempty"`
	// Nodes describe the vertex list for the Graph.
	//
	// See Node.
	Nodes []Node `json:"nodes,omitempty"`
}

func (g *Graph) Clone() *Graph {
	if g == nil {
		return nil
	}
	gCopy := &Graph{
		Adjacency: make([]Adjacency, 0, len(g.Adjacency)),
		Edges:     make([]Edge, 0, len(g.Edges)),
		Nodes:     make([]Node, 0, len(g.Nodes)),
	}
	for _, a := range g.Adjacency {
		gCopy.Adjacency = append(gCopy.Adjacency, a.Clone())
	}
	for _, e := range g.Edges {
		gCopy.Edges = append(gCopy.Edges, e.Clone())
	}
	for _, n := range g.Nodes {
		gCopy.Nodes = append(gCopy.Nodes, n.Clone())
	}
	return gCopy
}

// MarshalJSON implements deterministic marshaling of Graph to JSON.
func (g *Graph) MarshalJSON() ([]byte, error) {
	if g == nil {
		return []byte("null"), nil
	}
	gCopy := g.Clone()
	slices.SortStableFunc(gCopy.Adjacency, func(a, b Adjacency) int { return strings.Compare(a.Key, b.Key) })
	slices.SortStableFunc(gCopy.Edges, func(a, b Edge) int { return strings.Compare(a.Key, b.Key) })
	slices.SortStableFunc(gCopy.Nodes, func(a, b Node) int { return strings.Compare(a.Key, b.Key) })

	var b bytes.Buffer
	b.WriteByte('{')
	var comma bool
	if len(gCopy.Adjacency) > 0 {
		comma = true
		b.WriteString(`"adjacency":`)
		data, err := json.Marshal(gCopy.Adjacency)
		if err != nil {
			return nil, err
		}
		b.Write(data)
	}
	if len(gCopy.Edges) > 0 {
		if comma {
			b.WriteByte(',')
		}
		b.WriteString(`"edges":`)
		data, err := json.Marshal(gCopy.Edges)
		if err != nil {
			return nil, err
		}
		b.Write(data)
	}
	if len(gCopy.Nodes) > 0 {
		if comma {
			b.WriteByte(',')
		}
		comma = true
		b.WriteString(`"nodes":`)
		data, err := json.Marshal(gCopy.Nodes)
		if err != nil {
			return nil, err
		}
		b.Write(data)
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

var (
	// ErrKey is returned when a Key is not found in a given context.
	//
	// Examples:
	//
	//	* SrcField is not defined in the given Op.
	//
	ErrKey = errors.New("key error")
	// ErrType is returned when a Type is not expected in a given context.
	//
	// Examples:
	//
	//	* An integer is passed to a StringOp.
	//
	ErrType = errors.New("type error")
	// ErrGlom is returned when a GlomOp is not expected in a given context.
	//
	// Examples:
	//
	//	* GlomAppend is passed to an aritmetic op.
	//
	ErrGlom = errors.New("glom error")
)
