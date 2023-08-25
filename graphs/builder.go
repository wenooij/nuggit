package graphs

import (
	"encoding/json"
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

func (b *Builder) Node(nodeType nuggit.OpKey, opts ...BuilderOption) string {
	b.once.Do(b.Reset)
	node := nuggit.Node{
		Op: nodeType,
	}
	var o builderOptions
	for _, fn := range opts {
		o.edgeCount = len(b.g.Edges) + len(o.edges) // Update for default edge naming.
		fn(&o)
	}
	if o.key == "" {
		o.key = fmt.Sprintf("%d", 1+len(b.g.Nodes))
	}
	if o.data != nil {
		if m, ok := o.data.(json.RawMessage); ok {
			node.Data = m
		} else {
			data, err := json.Marshal(o.data)
			if err != nil {
				// TODO(wes): Don't panic!
				panic(err)
			}
			node.Data = data
		}
	}
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
			Glom:     e.glom,
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

type BuilderOption func(b *builderOptions)

func Data(data any) BuilderOption {
	return func(b *builderOptions) {
		b.data = data
	}
}

func Key(k nuggit.Key) BuilderOption {
	return func(b *builderOptions) {
		b.key = k
	}
}

func Stage(s nuggit.StageKey) BuilderOption {
	return func(b *builderOptions) {
		b.stage = s
	}
}

func Edge(dst nuggit.Key, opts ...EdgeOption) BuilderOption {
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
	glom     nuggit.Glom
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

func Glom(op nuggit.Glom) EdgeOption {
	return func(o *edgeOptions) {
		o.glom = op
	}
}
