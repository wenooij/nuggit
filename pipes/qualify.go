package pipes

import (
	"fmt"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/integrity"
)

// Qualified returns a new index with all pipes qualified by digest.
func (i *Index) Qualified() (*Index, error) {
	var newIndex Index

	for nd, err := range i.Topo() {
		if err != nil {
			return nil, err
		}

		pipe, ok := i.Get(nd.GetName(), nd.GetDigest())
		if !ok {
			return nil, fmt.Errorf("failed to find pipe in index (%q)", nd)
		}

		qualified, err := newIndex.Qualify(pipe)
		if err != nil {
			return nil, err
		}

		digest, err := integrity.GetDigest(integrity.DummySpec{Spec: qualified})
		if err != nil {
			return nil, err
		}

		newIndex.Add(nd.GetName(), digest, qualified)
	}

	return &newIndex, nil
}

// Qualify replaces all pipe references with a valid digest.
//
// Qualify does not update the pipe digest in the index.
//
// Use Qualified to create a new qualified index.
func (idx *Index) Qualify(pipe nuggit.Pipe) (nuggit.Pipe, error) {
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
