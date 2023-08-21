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
	Adjacency map[nuggit.NodeKey]nuggit.Adjacency
	Edges     map[nuggit.EdgeKey]nuggit.Edge
	Nodes     map[nuggit.NodeKey]nuggit.Node
	Stages    map[nuggit.StageKey]nuggit.Stage
	stageMap  map[nuggit.NodeKey]nuggit.StageKey
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
	gg.initStageMap()
	return gg
}

func (g *Graph) initStageMap() {
	g.stageMap = make(map[nuggit.NodeKey]nuggit.StageKey, len(g.Nodes))
	for _, s := range g.Stages {
		for _, k := range s.Nodes {
			g.stageMap[k] = s.Key
		}
	}
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

// Delete removes the node from the graph and all edges.
// It returns the pruned edges.
func (g *Graph) Delete(k nuggit.NodeKey) []nuggit.Edge {
	a := g.Adjacency[k]
	delete(g.Adjacency, k)
	delete(g.Nodes, k)
	es := make([]nuggit.Edge, len(a.Edges))
	for _, e := range a.Edges {
		es = append(es, g.Edges[e])
		delete(g.Edges, e)
	}
	return es
}

func (g *Graph) Clone() *Graph {
	gg := &Graph{
		Adjacency: maps.Clone(g.Adjacency),
		Edges:     maps.Clone(g.Edges),
		Nodes:     maps.Clone(g.Nodes),
		Stages:    maps.Clone(g.Stages),
	}
	gg.initStageMap()
	return gg
}

// Stage returns the stage for the node key k.
func (g *Graph) Stage(k nuggit.NodeKey) nuggit.StageKey {
	return g.stageMap[k]
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

type Subgraph struct {
	Parent *Graph
	*Graph
	EdgesIn []nuggit.Edge
}

// Prune the given nodes from the graph and all edges.
// It returns the pruned edges.
func (g *Graph) Prune(keys []nuggit.NodeKey) *Subgraph {
	if len(keys) == 0 {
		return &Subgraph{Parent: g, Graph: g}
	}
	pruneKeys := make(map[nuggit.NodeKey]struct{})
	for _, k := range keys {
		pruneKeys[k] = struct{}{}
	}
	return g.SubgraphFunc(func(n nuggit.Node) bool { _, ok := pruneKeys[n.Key]; return !ok })
}

func (g *Graph) SubgraphFunc(fn func(nuggit.Node) bool) *Subgraph {
	var edges []nuggit.Edge
	subgraph := g.Clone()
	for _, n := range g.Nodes {
		if !fn(n) {
			es := subgraph.Delete(n.Key)
			edges = append(edges, es...)
		}
	}
	return &Subgraph{Parent: g, Graph: subgraph, EdgesIn: edges}
}

// StageGraph returns the subgraph of g for the given stage.
// It returns the pruned edges which enter the subgraph.
func (g *Graph) StageGraph(key nuggit.StageKey) *Subgraph {
	pruneKeys := make([]nuggit.NodeKey, 0, len(g.Nodes))
	for k := range g.Nodes {
		if g.Stage(k) != key {
			pruneKeys = append(pruneKeys, k)
		}
	}
	subgraph := g.Prune(pruneKeys)
	edgesIn := make([]nuggit.Edge, 0, len(subgraph.Edges))
	for _, e := range subgraph.Edges {
		if g.Stage(e.Dst) == key {
			edgesIn = append(edgesIn, e)
		}
	}
	subgraph.EdgesIn = edgesIn
	return subgraph
}
