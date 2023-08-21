package graphs

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/wenooij/nuggit"
)

type Builder struct {
	adjacency map[nuggit.Key][]nuggit.EdgeKey
	stages    map[nuggit.StageKey][]nuggit.Key
	edges     []nuggit.Edge
	nodes     map[nuggit.Key]nuggit.Node
	once      sync.Once
}

func (b *Builder) init() {
	b.adjacency = make(map[nuggit.Key][]nuggit.EdgeKey)
	b.stages = make(map[nuggit.StageKey][]nuggit.Key)
	b.nodes = make(map[nuggit.Key]nuggit.Node)
}

func (b *Builder) Node(nodeType nuggit.OpKey, opts ...BuilderOption) string {
	b.once.Do(b.init)
	node := nuggit.Node{
		Op: nodeType,
	}
	var o builderOptions
	for _, fn := range opts {
		o.edgeCount = len(b.edges) + len(o.edges) // Update for default edge naming.
		fn(&o)
	}
	if o.key == "" {
		o.key = fmt.Sprintf("%d", 1+len(b.nodes))
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
		b.stages[k] = append(b.stages[k], o.key)
	}
	node.Key = o.key
	b.nodes[o.key] = node
	for _, e := range o.edges {
		b.edges = append(b.edges, nuggit.Edge{
			Key:      e.key,
			Src:      o.key,
			Dst:      e.dst,
			SrcField: e.srcField,
			DstField: e.dstField,
			Glom:     e.glom,
		})
		b.adjacency[o.key] = append(b.adjacency[o.key], e.key)
	}
	return o.key
}

func (b *Builder) Build() *nuggit.Graph {
	adjacency := make([]nuggit.Adjacency, 0, len(b.adjacency))
	for k, es := range b.adjacency {
		adjacency = append(adjacency, nuggit.Adjacency{
			Key:   k,
			Edges: es,
		})
	}
	stages := make([]nuggit.Stage, 0, len(b.stages))
	for k, ns := range b.stages {
		stages = append(stages, nuggit.Stage{
			Key:   k,
			Nodes: ns,
		})
	}
	nodes := make([]nuggit.Node, 0, len(b.nodes))
	for _, step := range b.nodes {
		nodes = append(nodes, step)
	}
	// TODO(ajzaff): Sort unordered keys before building.
	return &nuggit.Graph{
		Adjacency: adjacency,
		Edges:     b.edges,
		Stages:    stages,
		Nodes:     nodes,
	}
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
