package views

import (
	"slices"

	"github.com/wenooij/nuggit"
)

func Clone(v nuggit.View) nuggit.View {
	copy := v
	copy.Columns = slices.Clone(v.Columns)
	return copy
}
