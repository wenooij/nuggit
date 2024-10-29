package pipes

import (
	"fmt"
	"iter"

	"github.com/wenooij/nuggit/integrity"
)

// Topo proceeds over the keys of the index topologically.
//
// If there are cycles in the topological order an error is passed to the yield function.
// Any other error is passed to yield as well.
//
// Topo does not attempt to Qualify names when no digest is present.
func (i *Index) Topo() iter.Seq2[integrity.NameDigest, error] {
	return func(yield func(integrity.NameDigest, error) bool) {
		deps := make(map[integrity.NameDigest]map[integrity.NameDigest]struct{})
		outDegree := make(map[integrity.NameDigest]int)

		for nd, pipe := range i.All() {
			for dep := range Deps(pipe) {
				ds := deps[dep]
				if ds == nil {
					ds = make(map[integrity.NameDigest]struct{})
					deps[dep] = ds
				}
				ds[nd] = struct{}{}
				outDegree[nd]++
			}
		}

		var queue []integrity.NameDigest
		for nd := range i.Keys() {
			if outDegree[nd] == 0 {
				queue = append(queue, nd)
			}
		}

		for len(queue) > 0 {
			nd := queue[0]
			queue = queue[1:]

			if !yield(nd, nil) {
				return
			}

			for dependent := range deps[nd] {
				if outDegree[dependent]--; outDegree[dependent] == 0 {
					queue = append(queue, dependent)
				}
			}
			delete(deps, nd)
		}

		if len(deps) > 0 {
			// Get any example dep cycle.
			var d1, d2 integrity.NameDigest
			for i := range deps {
				d1 = i
				for j := range deps {
					d2 = j
					break
				}
				break
			}
			yield(nil, fmt.Errorf("at least one cycle exists in pipes (%q -> %q)", d1, d2))
			return
		}
	}
}
