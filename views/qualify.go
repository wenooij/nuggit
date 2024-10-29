package views

import (
	"fmt"
	"slices"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/integrity"
	"github.com/wenooij/nuggit/pipes"
)

// Qualify adds digests to referenced pipes, where not present.
//
// Qualify fails if the pipes index does not have a unique entry for the given name
// or if the name@digest is malformed.
func Qualify(pipes *pipes.Index, view nuggit.View) (nuggit.View, error) {
	copyCols := slices.Clip(slices.Clone(view.Columns))
	for i, c := range copyCols {
		nameDigest, err := integrity.ParseNameDigest(c.Pipe)
		if err != nil {
			return nuggit.View{}, err
		}
		if nameDigest.GetDigest() == "" {
			digest, ok := pipes.GetUnique(nameDigest.GetName())
			if !ok {
				return nuggit.View{}, fmt.Errorf("no unique pipe found (%q)", nameDigest.GetName())
			}
			pipe, err := integrity.FormatString(integrity.KeyLit(nameDigest.GetName(), digest))
			if err != nil {
				return nuggit.View{}, err
			}
			copyCols[i].Pipe = pipe
		}
	}
	view.Columns = copyCols
	return view, nil
}
