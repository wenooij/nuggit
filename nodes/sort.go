package nodes

import (
	"strings"

	"github.com/wenooij/nuggit"
	"golang.org/x/exp/slices"
)

func Sort(nodes []nuggit.Node) {
	slices.SortFunc(nodes, func(a, b nuggit.Node) int {
		if a.Key == b.Key {
			return strings.Compare(a.Op, b.Op)
		}
		if a.Key < b.Key {
			return -1
		}
		return +1
	})
}
