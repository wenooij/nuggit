package packages

import (
	"iter"
	"maps"
	"slices"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/integrity"
)

// Iter produces an arbitrary iteration over resources in an expanded package.
func Iter(p nuggit.Package) iter.Seq2[*api.Resource, error] {
	return func(yield func(*api.Resource, error) bool) {
		for k, u := range p.Rules {
			r := new(api.Resource)
			r.Kind = api.KindRule
			r.Metadata = new(api.ResourceMetadata)
			r.SetName(k)
			r.ReplaceSpec(&u)
			if !yield(r, nil) {
				return
			}
		}
		for k, v := range p.Views {
			r := new(api.Resource)
			r.Kind = api.KindView
			r.Metadata = new(api.ResourceMetadata)
			r.SetName(k)
			r.ReplaceSpec(&v)
			if !yield(r, nil) {
				return
			}
		}
		uniqueLabels := make(map[string]map[string]struct{})
		for k, labels := range p.AdditionalLabels {
			mappedLabels := make(map[string]struct{})
			for _, label := range labels {
				mappedLabels[label] = struct{}{}
			}
			uniqueLabels[k] = mappedLabels
		}

		for k, e := range p.Pipes {
			r := new(api.Resource)
			r.Kind = api.KindPipe
			r.Metadata = new(api.ResourceMetadata)
			r.SetName(k)
			r.ReplaceSpec(&e)
			labels := slices.Sorted(maps.Keys(uniqueLabels[k]))
			for _, label := range labels {
				r.Metadata.Labels = append(r.Metadata.Labels, label)
			}
			if !yield(r, nil) {
				return
			}
		}
	}
}

// Keys produces an arbitrary iteration over resource name@digest keys in a package's expanded resources.
//
// Use Iter to construct and iterate over the actual resources.
func Keys(p nuggit.Package) iter.Seq2[integrity.NameDigest, error] {
	return func(yield func(integrity.NameDigest, error) bool) {
		yieldKey := func(name string, spec any) bool {
			digest, err := integrity.GetDigest(integrity.DummySpec{Spec: spec})
			if err != nil {
				yield(nil, err)
				return false
			}
			if !yield(integrity.KeyLit(name, digest), nil) {
				return false
			}
			return true
		}
		for k, u := range p.Rules {
			if !yieldKey(k, u) {
				return
			}
		}
		for k, v := range p.Views {
			if !yieldKey(k, v) {
				return
			}
		}
		for k, e := range p.Pipes {
			if !yieldKey(k, e) {
				return
			}
		}
	}
}
