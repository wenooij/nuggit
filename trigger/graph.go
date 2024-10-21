package trigger

import (
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

	compareNameDigests := func(a, b api.NameDigest) int {
		if cmp := strings.Compare(a.Digest, b.Digest); cmp != 0 {
			return cmp
		}
		return strings.Compare(a.Name, b.Name)
	}

	queue := make([]*graphNode, 0, len(g.root.next))

	enqueueSortedNodes := func(n *graphNode) {
		sortedKeys := slices.SortedFunc(maps.Keys(n.next), compareNameDigests)
		for _, k := range sortedKeys {
			queue = append(queue, n.next[k])
		}
	}

	enqueueSortedNodes(g.root)

	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		if !yield(n) {
			return
		}
		enqueueSortedNodes(n)
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
	nd, err := api.NewNameDigest(&a)
	if err != nil {
		return err
	}
	next, found := n.next[nd]
	if !found { // Add new child.
		next = &graphNode{action: a}
		if n.next == nil {
			n.next = make(map[api.NameDigest]*graphNode, 2)
		}
		n.next[nd] = next
	}
	return next.add(pipe, actions[1:], exchangeAdded)
}
