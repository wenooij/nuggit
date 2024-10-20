package trigger

import (
	"iter"
	"maps"
	"slices"
	"strings"

	"github.com/wenooij/nuggit/api"
)

type graph struct {
	root *graphNode
}

func newGraph() *graph {
	g := &graph{}
	g.root = &graphNode{g: g}
	return g
}

func (g *graph) add(pipe *api.Pipe) error {
	nd, err := api.NewNameDigest(pipe)
	if err != nil {
		return err
	}
	return g.root.add(nd, pipe.GetActions(), false /* = exchangeAdded */)
}

func (g *graph) Len() int {
	if g == nil || g.root == nil {
		return 0
	}
	return len(g.root.next)
}

// consistentTopoIter iterates over all graph nodes in a consistent way.
func (g *graph) consistentTopoIter(yield func(*graphNode) bool) {
	if g == nil || g.root == nil {
		return
	}

	sortedNameDigests := func(keys iter.Seq[api.NameDigest]) []api.NameDigest {
		return slices.SortedFunc(keys, func(a, b api.NameDigest) int {
			if cmp := strings.Compare(a.Digest, b.Digest); cmp != 0 {
				return cmp
			}
			return strings.Compare(a.Name, b.Name)
		})
	}

	nodes := maps.Clone(g.root.next)
	nds := sortedNameDigests(maps.Keys(nodes))

	for len(nds) > 0 {
		nd := nds[0]
		nds = nds[1:]
		n := nodes[nd]
		if !yield(n) {
			return
		}
		nds = append(nds, sortedNameDigests(maps.Keys(n.next))...)
	}
}

type graphNode struct {
	g      *graph
	action api.Action
	next   map[api.NameDigest]*graphNode
}

func (n *graphNode) add(pipe api.NameDigest, actions []api.Action, exchangeAdded bool) error {
	if len(actions) == 0 {
		if !exchangeAdded { // Add exchange node here.
			return n.add(pipe, []api.Action{{Action: api.ActionExchange, Spec: api.ExchangeAction{Pipe: pipe.String()}}}, true /* = exchangeAdded */)
		}
		return nil
	}
	a := actions[0]
	nd, err := api.NewNameDigest(a)
	if err != nil {
		return err
	}
	n, found := n.next[nd]
	if !found { // Add new child.
		n = &graphNode{action: a}
		n.next[nd] = n
	}
	return n.add(pipe, actions[1:], exchangeAdded)
}
