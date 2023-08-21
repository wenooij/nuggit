package nodes

import (
	"github.com/wenooij/nuggit"
)

func Keys(nodes []nuggit.Node) []nuggit.NodeKey {
	keys := make([]nuggit.NodeKey, 0, len(nodes))
	for _, n := range nodes {
		keys = append(keys, n.Key)
	}
	return keys
}
