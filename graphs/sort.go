package graphs

import (
	"fmt"

	"github.com/wenooij/nuggit"
)

const (
	marked = 1
	done   = 2
)

// Sort nodes using a topological sort.
func (g *Graph) Sort(keys []nuggit.NodeKey) error {
	res := make([]nuggit.NodeKey, 0, len(keys))
	marks := make(map[nuggit.NodeKey]int, len(keys))
	for hasChanges := true; hasChanges; {
		hasChanges = false
		for _, k := range keys {
			change, err := g.visit(marks, &res, k)
			if err != nil {
				return err
			}
			if change {
				hasChanges = true
			}
		}
	}
	copy(keys, res)
	return nil
}

func (g *Graph) visit(marks map[nuggit.NodeKey]int, out *[]nuggit.NodeKey, k nuggit.NodeKey) (bool, error) {
	mark := marks[k]
	if mark == done {
		return false, nil
	}
	if mark == marked {
		return false, fmt.Errorf("graph has a cycle at: %q", k)
	}
	marks[k] = marked

	for _, e := range g.Adjacency[k].Edges {
		if _, err := g.visit(marks, out, g.Edges[e].Dst); err != nil {
			return false, err
		}
	}

	marks[k] = done
	*out = append(*out, k)
	return true, nil
}
