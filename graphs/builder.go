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

func (b *Builder) Reset() {
	b.g = FromGraph(&nuggit.Graph{})
}

func (b *Builder) Stage(stage nuggit.StageKey) {
	b.g.Stage = stage
}

func (b *Builder) Node(op nuggit.Op, opts ...NodeOption) string {
	b.once.Do(b.Reset)
	node := nuggit.Node{
		Op: op,
	}
	var o builderOptions
	for _, fn := range opts {
		o.edgeCount = len(b.g.Edges) + len(o.edges) // Update for default edge naming.
		fn(&o)
	}
	if o.key == "" {
		o.key = fmt.Sprintf("%d", 1+len(b.g.Nodes))
	}
	node.Data = o.data
	if k := o.stage; k != "" {
		b.Stage(k)
	}
	node.Key = o.key
	b.g.Nodes[o.key] = node
	for _, e := range o.edges {
		b.g.Edges[e.key] = nuggit.Edge{
			Key:      e.key,
			Src:      o.key,
			Dst:      e.dst,
			SrcField: e.srcField,
			DstField: e.dstField,
		}
		a := b.g.Adjacency[o.key]
		a.Key = o.key
		a.Edges = append(a.Edges, e.key)
		b.g.Adjacency[o.key] = a
	}
	return o.key
}

func (b *Builder) Build() *nuggit.Graph {
	return b.g.Graph()
}

type builderOptions struct {
	edgeCount int
	key       nuggit.Key
	data      any
	stage     nuggit.StageKey
	edges     []edgeOptions
}

type NodeOption func(b *builderOptions)

func Data(data any) NodeOption {
	return func(b *builderOptions) {
		b.data = data
	}
}

func Key(k nuggit.Key) NodeOption {
	return func(b *builderOptions) {
		b.key = k
	}
}

func Stage(s nuggit.StageKey) NodeOption {
	return func(b *builderOptions) {
		b.stage = s
	}
}

func Edge(dst nuggit.Key, opts ...EdgeOption) NodeOption {
	return func(b *builderOptions) {
		var o edgeOptions
		o.dst = dst
		for _, f := range opts {
			f(&o)
		}
		if o.key == "" {
			o.key = fmt.Sprintf("e%d", 1+b.edgeCount)
		}
		b.edges = append(b.edges, o)
	}
}

type edgeOptions struct {
	key      nuggit.EdgeKey
	dst      nuggit.EdgeKey
	dstField nuggit.FieldKey
	srcField nuggit.FieldKey
}

type EdgeOption func(*edgeOptions)

func EdgeKey(key string) EdgeOption {
	return func(o *edgeOptions) {
		o.key = key
	}
}

func SrcField(k nuggit.FieldKey) EdgeOption {
	return func(o *edgeOptions) {
		o.srcField = k
	}
}

func DstField(k nuggit.FieldKey) EdgeOption {
	return func(o *edgeOptions) {
		o.dstField = k
	}
}
