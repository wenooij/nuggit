// Package graphs implements utility functions on Nuggit Graphs.
package graphs

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/resources"
	"github.com/wenooij/nuggit/vars"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// Graph implements a Graph structure that supports efficient for Graph Key lookup.
//
// See nuggit.Graph.
type Graph struct {
	Stage     string
	Adjacency map[string]Adjacency
	Edges     map[string]nuggit.Edge
	Nodes     map[string]nuggit.Node
}

func NewGraph() *Graph {
	return &Graph{
		Adjacency: make(map[string]Adjacency),
		Nodes:     make(map[string]nuggit.Node),
		Edges:     make(map[string]nuggit.Edge),
	}
}

// FromGraph loads a Graph from a Nuggit Graph spec.
func FromGraph(g *nuggit.Graph) *Graph {
	if g == nil {
		return NewGraph()
	}
	gg := &Graph{
		Adjacency: make(map[string]Adjacency, len(g.Adjacency)),
		Nodes:     make(map[string]nuggit.Node, len(g.Nodes)),
		Edges:     make(map[string]nuggit.Edge, len(g.Edges)),
	}
	for _, a := range g.Adjacency {
		gg.Adjacency[a.Key] = FromAdjacency(a)
	}
	for _, n := range g.Nodes {
		gg.Nodes[n.Key] = n
	}
	for _, e := range g.Edges {
		gg.Edges[e.Key] = e
	}
	return gg
}

// FromFile loads a Graph from a JSON file.
// The file may be either a JSON encoded Resource or Graph.
// Checksum integrity checks are ignored.
func FromFile(filename string) (*Graph, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var g *nuggit.Graph
	if r, err := resources.FromFile(filename); err == nil {
		if r.Kind != "Graph" {
			return nil, fmt.Errorf("unexpected Resource type: %q", r.Kind)
		}
		g = r.Spec.(*nuggit.Graph)
	} else {
		g := new(nuggit.Graph)
		if err := json.Unmarshal(data, &g); err != nil {
			return nil, err
		}
	}
	return FromGraph(g), nil
}

func (g *Graph) Clone() *Graph {
	return &Graph{
		Adjacency: maps.Clone(g.Adjacency),
		Edges:     maps.Clone(g.Edges),
		Nodes:     maps.Clone(g.Nodes),
	}
}

// Graph returns the deterministic canonical Graph.
func (g *Graph) Graph() *nuggit.Graph {
	gg := &nuggit.Graph{
		Adjacency: make([]nuggit.Adjacency, 0, len(g.Adjacency)),
		Edges:     make([]nuggit.Edge, 0, len(g.Edges)),
		Nodes:     make([]nuggit.Node, 0, len(g.Nodes)),
	}
	adjacencyKeys := maps.Keys(g.Adjacency)
	slices.Sort(adjacencyKeys)
	for _, src := range adjacencyKeys {
		gg.Adjacency = append(gg.Adjacency, g.Adjacency[src].Adjacency(src))
	}
	edgeKeys := maps.Keys(g.Edges)
	slices.Sort(edgeKeys)
	for _, e := range edgeKeys {
		gg.Edges = append(gg.Edges, g.Edges[e])
	}
	nodeKeys := maps.Keys(g.Nodes)
	slices.Sort(nodeKeys)
	for _, n := range nodeKeys {
		gg.Nodes = append(gg.Nodes, g.Nodes[n])
	}
	return gg
}

type Adjacency map[string]struct{}

func FromAdjacency(a nuggit.Adjacency) Adjacency {
	m := Adjacency{}
	for _, e := range a.Elems {
		m[e] = struct{}{}
	}
	return m
}

func (a Adjacency) Adjacency(src string) nuggit.Adjacency {
	edges := maps.Keys(a)
	sort.Strings(edges)
	return nuggit.Adjacency{Key: src, Elems: edges}
}

func (g *Graph) Var(name string) vars.Var {
	for k, n := range g.Nodes {
		if n.Op == "Var" && k == name {
			return GraphVar{g: g, v: k}
		}
	}
	return nil
}

type GraphVar struct {
	g *Graph
	v string
}

func (v GraphVar) SetDefault(x any) error {
	panic("not implemented")
}

func (v GraphVar) Set(x any) error {
	panic("not implemented")
}

func (v GraphVar) Get() (any, error) {
	panic("not implemented")
}
