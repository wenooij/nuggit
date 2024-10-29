package pipes

import (
	"fmt"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/integrity"
)

// Qualify replaces all instances of pipe actions with the unique pipe from the unique func.
//
// Qualify returns a new pipe or an error if the qualification failed.
//
// TODO:
// NB:
// Qualify must be called in topologically sorted order.
// Due to this issue, we currently don't use it in client APIs.
// For now we just flatten all pipes.
func Qualify(idx *Index, pipe nuggit.Pipe) (nuggit.Pipe, error) {
	pipeCopy := Clone(pipe)
	for _, a := range pipeCopy.Actions {
		if a.GetAction() != "pipe" {
			continue
		}
		name, digest := a.GetOrDefaultArg("name"), a.GetOrDefaultArg("digest")
		if digest != "" && !idx.Has(name, digest) {
			return nuggit.Pipe{}, fmt.Errorf("pipe not found in index (%q)", integrity.KeyLit(name, digest))
		}
		if !idx.HasName(name) {
			return nuggit.Pipe{}, fmt.Errorf("pipe not found in index (%q)", name)
		}
		digest, ok := idx.GetUnique(name)
		if !ok {
			return nuggit.Pipe{}, fmt.Errorf("index has conflicting entries (%q)", name)
		}
		a.SetDigest(digest)
	}
	return pipeCopy, nil
}
