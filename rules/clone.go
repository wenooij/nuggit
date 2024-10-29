package rules

import (
	"slices"

	"github.com/wenooij/nuggit"
)

func Clone(r nuggit.Rule) nuggit.Rule {
	copy := r
	copy.Labels = slices.Clone(r.Labels)
	return copy
}
