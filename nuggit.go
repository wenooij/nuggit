// Package nuggit provides a declarative API for information retrieval (IR).
package nuggit

import (
	"errors"
)

// Adjacency describes the elements that are adjacent to a given Key.
// Elems are Exchanges when used in part of a Plan.
type Adjacency struct {
	// Key describes the src Node.
	Key string `json:"key,omitempty"`
	// Elems lists all adjacent edges.
	Elems []string `json:"elems,omitempty"`
}

// Stage describes a sequence of op Nodes which are executed together.
type Stage struct {
	// Key is a unique id for the Stage.
	Key string `json:"key,omitempty"`
	// Nodes is an ordered list of the Nodes in this stage.
	Nodes []string `json:"nodes,omitempty"`
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
	Key string `json:"key,omitempty"`
	// Src graph for exchanges. Only required when different from DstGraph.
	SrcGraph string `json:"src_graph,omitempty"`
	// Dst graph for exchanges. Only required when different from SrcGraph.
	DstGraph string `json:"dst_graph,omitempty"`
	// Src describes the starting Node.
	Src string `json:"src,omitempty"`
	// Dst describes the terminal Node.
	Dst string `json:"dst,omitempty"`
	// SrcField describes the field within Src, if set.
	//
	// See the doc for specific Ops to see which fields are defined.
	SrcField string `json:"src_field,omitempty"`
	// DstField describes the field within Dst, if set.
	//
	// See the doc for specific Ops to see which fields are defined.
	DstField string `json:"dst_field,omitempty"`
	// Data specifies arbitrary JSON data to attach to this Edge.
	//
	// The Op runtime may use Data to change the semantics of the connection.
	Data any `json:"data,omitempty"`
}

// Node describes a node in a Graph.
// Data provides additional configuration to the Op.
//
// See Op documentation to see which Ops are defined.
type Node struct {
	Key string `json:"key,omitempty"`
	// Op specifies
	Op Op `json:"op,omitempty"`
	// Data specifies arbitrary data to attach to this Edge.
	//
	// Data can be used to alter the behavior of the Op.
	Data any `json:"data,omitempty"`
}

// Graph describes a declarative program DAG.
type Graph struct {
	// Key is the unique identifier for the Graph.
	Key string `json:"key,omitempty"`
	// Stage is the name of the stage for this Graph.
	Stage string `json:"stage,omitempty"`
	// Adjacency describes the adjacency list for the Graph.
	// Adjacencies relate NodeKeys to the ordered lists of EdgeKeys sourced there.
	Adjacency []Adjacency `json:"adjacency,omitempty"`
	// Edges describe the edge list for the Graph.
	Edges []Edge `json:"edges,omitempty"`
	// Nodes describe the vertex list for the Graph.
	Nodes []Node `json:"nodes,omitempty"`
}

// Plan represents a list of interconnected Graphs defining a multistage program.
type Plan struct {
	Graphs []*Graph `json:"graphs,omitempty"`
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
)
