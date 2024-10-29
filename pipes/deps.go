package pipes

import (
	"iter"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/integrity"
)

func Deps(p nuggit.Pipe) iter.Seq[integrity.NameDigest] {
	return func(yield func(integrity.NameDigest) bool) {
		for _, a := range p.Actions {
			if a.GetAction() != "pipe" {
				continue
			}
			key := integrity.KeyLit(a.GetOrDefaultArg("name"), a.GetOrDefaultArg("digest"))
			if !yield(key) {
				return
			}
		}
	}
}
