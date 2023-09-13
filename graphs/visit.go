package graphs

import (
	"fmt"

	"github.com/wenooij/nuggit"
)

func Visit(g *Graph, visitFn func(k nuggit.NodeKey) error) error {
	marks := make(map[nuggit.NodeKey]int, len(g.Nodes))
	for hasChanges := true; hasChanges; {
		hasChanges = false
		for k := range g.Nodes {
			change, err := visit(g, marks, k, visitFn)
			if err != nil {
				return err
			}
			if change {
				hasChanges = true
			}
		}
	}
	return nil
}

func visit(g *Graph, marks map[nuggit.NodeKey]int, k nuggit.NodeKey, visitFn func(k nuggit.NodeKey) error) (bool, error) {
	const (
		marked = 1
		done   = 2
	)
	mark := marks[k]
	if mark == done {
		return false, nil
	}
	if mark == marked {
		return false, fmt.Errorf("graph has a cycle at: %q", k)
	}
	marks[k] = marked

	for _, e := range g.Adjacency[k].Edges {
		if _, err := visit(g, marks, g.Edges[e].Dst, visitFn); err != nil {
			return false, err
		}
	}

	marks[k] = done
	if err := visitFn(k); err != nil {
		return false, err
	}
	return true, nil
}
