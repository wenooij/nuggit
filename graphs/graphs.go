// Package graphs implements utility functions on Nuggit Graphs.
package graphs

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/resources"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// Graph implements a Graph structure that supports efficient for Graph Key lookup.
//
// See nuggit.Graph.
type Graph struct {
	Adjacency map[nuggit.Key]nuggit.Adjacency
	Edges     map[nuggit.EdgeKey]nuggit.Edge
	Nodes     map[nuggit.Key]nuggit.Node
	Stages    map[nuggit.StageKey]nuggit.Stage
}

// FromGraph loads a Graph from a Nuggit Graph spec.
func FromGraph(g *nuggit.Graph) *Graph {
	if g == nil {
		return nil
	}
	gg := &Graph{
		Adjacency: make(map[nuggit.Key]nuggit.Adjacency, len(g.Adjacency)),
		Nodes:     make(map[nuggit.Key]nuggit.Node, len(g.Nodes)),
		Edges:     make(map[nuggit.EdgeKey]nuggit.Edge, len(g.Edges)),
		Stages:    make(map[nuggit.StageKey]nuggit.Stage, len(g.Stages)),
	}
	for _, a := range g.Adjacency {
		gg.Adjacency[a.Key] = a.Clone()
	}
	for _, n := range g.Nodes {
		gg.Nodes[n.Key] = n.Clone()
	}
	for _, e := range g.Edges {
		gg.Edges[e.Key] = e.Clone()
	}
	for _, s := range g.Stages {
		gg.Stages[s.Key] = s.Clone()
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

func (g *Graph) Graph() *nuggit.Graph {
	gg := &nuggit.Graph{
		Adjacency: make([]nuggit.Adjacency, 0, len(g.Adjacency)),
		Edges:     make([]nuggit.Edge, 0, len(g.Edges)),
		Nodes:     make([]nuggit.Node, 0, len(g.Nodes)),
		Stages:    make([]nuggit.Stage, 0, len(g.Adjacency)),
	}
	adjacencyKeys := maps.Keys(g.Adjacency)
	slices.Sort(adjacencyKeys)
	for _, a := range adjacencyKeys {
		gg.Adjacency = append(gg.Adjacency, g.Adjacency[a])
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
	stageKeys := maps.Keys(g.Stages)
	slices.Sort(stageKeys)
	for _, s := range stageKeys {
		gg.Stages = append(gg.Stages, g.Stages[s])
	}
	return gg
}
