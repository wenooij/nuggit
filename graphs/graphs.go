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
	Stage     nuggit.StageKey
	Adjacency map[nuggit.NodeKey]Adjacency
	Edges     map[nuggit.EdgeKey]nuggit.Edge
	Nodes     map[nuggit.NodeKey]nuggit.Node
}

// FromGraph loads a Graph from a Nuggit Graph spec.
func FromGraph(g *nuggit.Graph) *Graph {
	gg := &Graph{
		Adjacency: make(map[nuggit.Key]Adjacency, len(g.Adjacency)),
		Nodes:     make(map[nuggit.Key]nuggit.Node, len(g.Nodes)),
		Edges:     make(map[nuggit.EdgeKey]nuggit.Edge, len(g.Edges)),
	}
	if g == nil {
		return gg
	}
	for _, a := range g.Adjacency {
		gg.Adjacency[a.Key] = FromAdjacency(a)
	}
	for _, n := range g.Nodes {
		gg.Nodes[n.Key] = n.Clone()
	}
	for _, e := range g.Edges {
		gg.Edges[e.Key] = e.Clone()
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

type Adjacency map[nuggit.EdgeKey]struct{}

func FromAdjacency(a nuggit.Adjacency) Adjacency {
	m := Adjacency{}
	for _, e := range a.Edges {
		m[e] = struct{}{}
	}
	return m
}

func (a Adjacency) Adjacency(src nuggit.NodeKey) nuggit.Adjacency {
	edges := maps.Keys(a)
	sort.Strings(edges)
	return nuggit.Adjacency{Key: src, Edges: edges}
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
	v nuggit.NodeKey
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
