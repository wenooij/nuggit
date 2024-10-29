package resources

import (
	"iter"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/integrity"
	"github.com/wenooij/nuggit/pipes"
	"github.com/wenooij/nuggit/views"
)

func (i *Index) Deps(r *api.Resource) iter.Seq[integrity.NameDigest] {
	switch r.GetKind() {
	case api.KindPipe:
		return pipes.Deps(*r.GetPipe())
	case api.KindView:
		return views.Deps(*r.GetView())
	default:
		return func(yield func(integrity.NameDigest) bool) {}
	}
}
