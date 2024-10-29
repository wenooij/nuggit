package resources

import (
	"slices"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/pipes"
	"github.com/wenooij/nuggit/rules"
	"github.com/wenooij/nuggit/views"
)

func Clone(r *api.Resource) *api.Resource {
	copy := *r
	copy.Metadata = CloneMetadata(r.Metadata)
	switch r.Kind {
	case api.KindPipe:
		pipe := pipes.Clone(*r.GetPipe())
		copy.Spec = &pipe
	case api.KindView:
		view := views.Clone(*r.GetView())
		copy.Spec = &view
	case api.KindRule:
		rule := rules.Clone(*r.GetRule())
		copy.Spec = &rule
	default:
		copy.Spec = nil
	}
	return &copy
}

func CloneMetadata(m *api.ResourceMetadata) *api.ResourceMetadata {
	copy := *m
	copy.Labels = slices.Clone(m.Labels)
	return &copy
}
