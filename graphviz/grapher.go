//go:build !windows

// TODO(wes): Provide grapher_windows.go for windows support.

package graphviz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	"github.com/goccy/go-graphviz/gvc"
	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/graphs"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type Grapher struct {
	*graphs.Graph
}

func (g *Grapher) Graphviz() *graphviz.Graphviz {
	gviz := graphviz.New()
	gviz.SetLayout(graphviz.DOT)
	gviz.SetRenderer(graphviz.SVG, &gvc.ImageRenderer{})
	return gviz
}

func (g *Grapher) CGraph(gviz *graphviz.Graphviz) (*cgraph.Graph, error) {
	graph, err := gviz.Graph()
	if err != nil {
		return nil, err
	}
	if g.Graph == nil {
		return graph, nil
	}

	nodes := maps.Keys(g.Nodes)
	geaphNodes := make(map[string]*cgraph.Node, len(nodes))
	slices.SortFunc(nodes, func(a, b nuggit.Key) int { return strings.Compare(a, b) })
	for _, k := range nodes {
		node := g.Nodes[k]
		n, err := graph.CreateNode(k)
		if err != nil {
			return nil, err
		}
		geaphNodes[k] = n
		n.SetShape(cgraph.BoxShape)

		var sb strings.Builder
		if node.Op == "" {
			fmt.Fprintf(&sb, "%s\\l", node.Op)
		} else {
			fmt.Fprintf(&sb, "%s(%s)\\l", node.Op, k)
		}
		es := g.Adjacency[k].Edges
		if len(es) > 0 {
			fmt.Fprintf(&sb, "Edges:\\l")
		}
		for _, e := range es {
			edge := g.Edges[e]

			fmt.Fprintf(&sb, "&nbsp;&nbsp;%s", edge.Key)

			srcField, dstField := edge.SrcField, edge.DstField
			glom := edge.Glom
			if srcField == "" && dstField == "" && glom == nuggit.GlomUndefined {
				fmt.Fprintf(&sb, "\\l")
				continue
			}
			if srcField == "" {
				srcField = "*"
			}
			if dstField == "" {
				dstField = "*"
			}
			var glomStr string
			if glom != nuggit.GlomUndefined {
				glomStr = fmt.Sprintf("[%s]", glom.String())
			}
			fmt.Fprintf(&sb, ": %s %s-> %s\\l", srcField, glomStr, dstField)
		}
		if node.Data != nil {
			fmt.Fprintf(&sb, "Data:\\l")
			data, err := json.MarshalIndent(node.Data, "", "  ")
			if err != nil {
				return nil, err
			}
			dataStr := linewrapMax(string(data))
			fmt.Fprintf(&sb, "&nbsp;&nbsp;%s\\l", dataStr)
		}
		n.SetLabel(sb.String())
	}

	edges := maps.Keys(g.Edges)
	slices.SortFunc(edges, func(a, b nuggit.EdgeKey) int { return strings.Compare(a, b) })
	for _, k := range edges {
		edge := g.Edges[k]
		e, err := graph.CreateEdge(edge.Key, geaphNodes[edge.Src], geaphNodes[edge.Dst])
		if err != nil {
			return nil, err
		}
		e.SetLabel(edge.Key)
	}

	return graph, nil
}

func (g *Grapher) DOT() ([]byte, error) {
	gviz := g.Graphviz()
	graph, err := g.CGraph(gviz)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := gviz.Render(graph, graphviz.XDOT, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (g *Grapher) SVG() ([]byte, error) {
	gviz := g.Graphviz()
	graph, err := g.CGraph(gviz)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := gviz.Render(graph, graphviz.SVG, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func linewrapMax(s string) string {
	const n = 512
	const w = 60

	var sb strings.Builder
	sb.Grow(len(s))
	var ct int
	for _, r := range s {
		if r == '\n' {
			sb.WriteString("\\l&nbsp;&nbsp;")
			ct = 0
			continue
		}
		sb.WriteRune(r)
		if ct > 0 && ct%w == 0 {
			sb.WriteString("\\l&nbsp;&nbsp;")
		}
		ct++
	}

	s = sb.String()
	if len(s) > n {
		return fmt.Sprintf("%s [omitted %d bytes]", s[:n], len(s)-n)
	}
	return s
}
