package graphs

import (
	"github.com/wenooij/nuggit"
)

// Sort nodes using a topological sort.
func Sort(g *Graph, keys []nuggit.NodeKey) error {
	visitKeys := make(map[nuggit.NodeKey]struct{})
	for _, k := range keys {
		visitKeys[k] = struct{}{}
	}
	res := make([]nuggit.NodeKey, 0, len(keys))
	if err := Visit(g, func(k nuggit.NodeKey) error {
		if _, ok := visitKeys[k]; ok {
			res = append(res, k)
		}
		return nil
	}); err != nil {
		return err
	}
	copy(keys, res)
	return nil
}
