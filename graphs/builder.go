package graphs

import (
	"fmt"
	"sync"

	"github.com/wenooij/nuggit"
)

type Builder struct {
	g    *Graph
	once sync.Once
}

func (b *Builder) Init() {
	b.g = FromGraph(&nuggit.Graph{})
}

func (b *Builder) Reset(g *Graph) {
	b.g = g
}

func (b *Builder) Stage(stage string) {
	b.once.Do(b.Init)
	b.g.Stage = stage
}

func (b *Builder) Node(op nuggit.Op, opts ...NodeOption) string {
	b.once.Do(b.Init)
	key := b.NextNodeKey()
	b.Insert(key, op, nil)
	for _, o := range opts {
		key = o(b, key)
	}
	return key
}

func (b *Builder) NextNodeKey() string {
	return fmt.Sprintf("%d", 1+len(b.g.Nodes))
}

func (b *Builder) NextEdgeKey() string {
	return fmt.Sprintf("e%d", 1+len(b.g.Edges))
}

// Delete removes the node from the graph and all edges.
// It returns the pruned node and edges.
func (b *Builder) Delete(k string) (nuggit.Node, []nuggit.Edge, bool) {
	if oldNode, ok := b.g.Nodes[k]; ok {
		a := b.g.Adjacency[k]
		oldEdges := make([]nuggit.Edge, 0, len(a))
		for e := range a {
			oldEdge, _ := b.DeleteEdge(e)
			oldEdges = append(oldEdges, oldEdge)
		}
		delete(b.g.Adjacency, k)
		delete(b.g.Nodes, k)
		return oldNode, oldEdges, true
	}
	return nuggit.Node{}, nil, false
}

func (b *Builder) deleteAdjacency(src, edge string) (removed bool) {
	if a := b.g.Adjacency[src]; a != nil {
		_, removed = a[edge]
		delete(a, edge)
		return removed
	}
	return false
}

func (b *Builder) DeleteEdge(k string) (nuggit.Edge, bool) {
	if oldEdge, ok := b.g.Edges[k]; ok {
		delete(b.g.Edges, k)
		b.deleteAdjacency(oldEdge.Src, oldEdge.Key)
		return oldEdge, true
	}
	return nuggit.Edge{}, false
}

func (b *Builder) insertAdjacency(src, edge string) (replaced bool) {
	a := b.g.Adjacency[src]
	if a == nil {
		a = make(Adjacency)
		b.g.Adjacency[src] = a
	}
	_, replaced = a[edge]
	a[edge] = struct{}{}
	return replaced
}

func (b *Builder) InsertEdge(key, dst, src, dstField, srcField string, data any) (oldEdge nuggit.Edge, replaced bool) {
	oldEdge, replaced = b.g.Edges[key]
	newEdge := nuggit.Edge{
		Key:      key,
		Src:      src,
		Dst:      dst,
		SrcField: srcField,
		DstField: dstField,
		Data:     data,
	}
	b.g.Edges[key] = newEdge
	b.insertAdjacency(src, key)
	return oldEdge, replaced
}

func (b *Builder) Insert(key, op string, data any) (oldNode nuggit.Node, oldEdges []nuggit.Edge, replaced bool) {
	oldNode, oldEdges, replaced = b.Delete(key)
	newNode := nuggit.Node{
		Key:  key,
		Op:   op,
		Data: data,
	}
	b.g.Nodes[key] = newNode
	return oldNode, oldEdges, replaced
}

func (b *Builder) Rename(oldKey, newKey string) (renamed bool) {
	if _, ok := b.g.Nodes[newKey]; ok {
		return false
	}
	if n, ok := b.g.Nodes[oldKey]; ok {
		n.Key = newKey
		b.g.Nodes[newKey] = n
		b.g.Adjacency[newKey] = b.g.Adjacency[oldKey]
		delete(b.g.Nodes, oldKey)
		delete(b.g.Adjacency, oldKey)
		return true
	}
	return false
}

func (b *Builder) RenameEdge(oldKey, newKey string) (renamed bool) {
	if _, ok := b.g.Edges[newKey]; ok {
		return false
	}
	if e, ok := b.g.Edges[oldKey]; ok {
		e.Key = newKey
		b.g.Edges[newKey] = e
		b.g.Adjacency[e.Src][newKey] = struct{}{}
		delete(b.g.Edges, oldKey)
		delete(b.g.Adjacency[e.Src], oldKey)
		return true
	}
	return false
}

func (b *Builder) InsertEdgeData(key string, data any) (oldData any, replaced bool) {
	if e, ok := b.g.Edges[key]; ok {
		oldData := e.Data
		e.Data = data
		b.g.Edges[key] = e
		return oldData, true
	}
	return nil, false
}

func (b *Builder) InsertData(key string, data any) (oldData any, replaced bool) {
	if n, ok := b.g.Nodes[key]; ok {
		oldData := n.Data
		n.Data = data
		b.g.Nodes[key] = n
		return oldData, true
	}
	return nil, false
}

func (b *Builder) Build() *nuggit.Graph {
	b.once.Do(b.Init)
	return b.g.Graph()
}

type NodeOption func(b *Builder, key string) (newKey string)

func Data(data any) NodeOption {
	return func(b *Builder, key string) (newKey string) {
		b.InsertData(key, data)
		return key
	}
}

func Key(newKey string) NodeOption {
	return func(b *Builder, oldKey string) (newKey string) {
		if b.Rename(oldKey, newKey) {
			return newKey
		}
		return oldKey
	}
}

func Stage(s string) NodeOption {
	return func(b *Builder, key string) (newKey string) {
		b.g.Stage = s
		return key
	}
}

type EdgeOption func(b *Builder, key string) (newKey string)

func Edge(dst string, opts ...EdgeOption) NodeOption {
	return func(b *Builder, src string) (newKey string) {
		key := b.NextEdgeKey()
		b.InsertEdge(key, dst, src, "", "", nil)
		for _, o := range opts {
			key = o(b, key)
		}
		return src
	}
}

func EdgeData(data any) EdgeOption {
	return func(b *Builder, key string) (newKey string) {
		b.InsertEdgeData(key, data)
		return key
	}
}

func EdgeKey(newKey string) EdgeOption {
	return func(b *Builder, oldKey string) (newKey string) {
		if b.RenameEdge(oldKey, newKey) {
			return newKey
		}
		return oldKey
	}
}

func SrcField(key string) EdgeOption {
	return func(b *Builder, key string) (newKey string) {
		if e, ok := b.g.Edges[key]; ok {
			e.SrcField = key
			b.g.Edges[key] = e
		}
		return key
	}
}

func DstField(k string) EdgeOption {
	return func(b *Builder, key string) (newKey string) {
		if e, ok := b.g.Edges[key]; ok {
			e.DstField = key
			b.g.Edges[key] = e
		}
		return key
	}
}
