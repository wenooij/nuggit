package resources

import (
	"fmt"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/integrity"
	"github.com/wenooij/nuggit/views"
)

// Qualify returns a new index with all resources qualified by digest.
//
// Currently Qualify only updates pipes.
func (i *Index) Qualified() (*Index, error) {
	var newIndex Index

	for nd, err := range i.Topo() {
		if err != nil {
			return nil, err
		}

		r, ok := i.Get(nd)
		if !ok {
			return nil, fmt.Errorf("failed to find resource in index (%q)", nd)
		}

		qualified, err := newIndex.Qualify(r)
		if err != nil {
			return nil, err
		}

		newIndex.Add(qualified)
	}

	return &newIndex, nil
}

// Qualify replaces all pipe references with a valid digest
// returning a new Resource.
//
// Qualify does not update the digest of r in the index.
//
// Use Qualified to create a new qualified index.
func (idx *Index) Qualify(r *api.Resource) (*api.Resource, error) {
	switch r.Kind {
	case api.KindPipe:
		resourceCopy := Clone(r)
		qualified, err := idx.Pipes().Qualify(*r.GetPipe())
		if err != nil {
			return nil, err
		}
		resourceCopy.ReplaceSpec(&qualified)
		integrity.SetDigest(resourceCopy)
		return resourceCopy, nil
	case api.KindView:
		resourceCopy := Clone(r)
		qualified, err := views.Qualify(idx.Pipes(), *r.GetView())
		if err != nil {
			return nil, err
		}
		resourceCopy.ReplaceSpec(&qualified)
		integrity.SetDigest(resourceCopy)
		return resourceCopy, nil
	case api.KindRule: // No qualification required for rules currently.
		return r, nil
	default:
		return r, nil
	}
}
