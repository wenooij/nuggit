package views

import (
	"iter"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/integrity"
)

func Deps(v nuggit.View) iter.Seq[integrity.NameDigest] {
	return func(yield func(integrity.NameDigest) bool) {
		for _, c := range v.Columns {
			nd, err := integrity.ParseNameDigest(c.Pipe)
			if err != nil {
				continue // Skip parsing errors.
			}
			if !yield(nd) {
				break
			}
		}
	}
}
