package pipes

import (
	"maps"

	"github.com/wenooij/nuggit"
)

func Clone(pipe nuggit.Pipe) nuggit.Pipe {
	copy := pipe
	if pipe.Actions == nil {
		return copy
	}
	actions := make([]nuggit.Action, len(pipe.Actions))
	for i, a := range pipe.Actions {
		actions[i] = maps.Clone(a)
	}
	copy.Actions = actions
	return copy
}
