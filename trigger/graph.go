package trigger

import (
	"maps"
	"slices"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/integrity"
)

type graph struct {
	root *graphNode
}

func newGraph() *graph {
	g := &graph{}
	g.root = &graphNode{g: g}
	return g
}

func (g *graph) add(nameDigest integrity.NameDigest, pipe nuggit.Pipe, actions []nuggit.Action) error {
	return g.root.add(nameDigest, pipe, actions, false /* = exchangeAdded */)
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

	queue := make([]*graphNode, 0, len(g.root.next))

	enqueueSortedNodes := func(n *graphNode) {
		sortedKeys := slices.Sorted(maps.Keys(n.next))
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
	action nuggit.Action
	next   map[string]*graphNode
}

func (n *graphNode) add(nameDigest integrity.NameDigest, pipe nuggit.Pipe, actions []nuggit.Action, exchangeAdded bool) error {
	if len(actions) == 0 {
		if !exchangeAdded { // Add exchange node here.
			// TODO: Add actions package to create this?
			action := nuggit.Action{
				"action": "exchange",
				"name":   nameDigest.GetName(),
				"digest": nameDigest.GetDigest(),
			}
			if scalar := pipe.Point.Scalar; scalar != nuggit.Bytes {
				action.SetOrDefault("scalar", scalar)
			}
			return n.add(nil, pipe, []nuggit.Action{action}, true /* = exchangeAdded */)
		}
		return nil
	}
	a := actions[0]
	digest, err := integrity.GetDigest(a)
	if err != nil {
		return err
	}
	next, found := n.next[digest]
	if !found { // Add new child.
		next = &graphNode{action: a}
		if n.next == nil {
			n.next = make(map[string]*graphNode, 2)
		}
		n.next[digest] = next
	}
	return next.add(nameDigest, pipe, actions[1:], exchangeAdded)
}
