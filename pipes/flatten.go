package pipes

import (
	"fmt"
	"slices"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/status"
)

// Flatten recursively replaces all pipe actions with their definitions
// returning a new Pipe or an error if the process failed.
// The flattened pipe is fully hermetric, making no references to other pipes.
// If the given pipe definition is not present in referencedPipes a ErrInvalidArgument
// error is returned.
//
// NOTE: The returned pipe will have a different digest than the input pipe.
//
// TODO: check the digests of pipes in referencedPipes.
func Flatten(referencedPipes map[api.NameDigest]*api.Pipe, pipe *api.Pipe) (*api.Pipe, error) {
	actions := slices.Clone(pipe.GetActions())
	for i := 0; i < len(actions); {
		a := actions[i]
		if a.GetAction() != "pipe" {
			i++
			continue
		}
		pipe := a.GetNameDigestArg()
		referencedPipe, ok := referencedPipes[pipe]
		if !ok {
			return nil, fmt.Errorf("referenced pipe not found (%q): %w", &pipe, status.ErrInvalidArgument)
		}
		actions = slices.Insert(slices.Delete(actions, i, i+1), i, referencedPipe.GetActions()...)
	}
	pipe = &api.Pipe{
		Actions: actions,
		Point:   pipe.GetPoint(),
		NameDigest: api.NameDigest{
			Name: pipe.GetName(),
		},
	}
	nameDigest, err := api.NewNameDigest(pipe)
	if err != nil {
		return nil, err
	}
	pipe.SetNameDigest(nameDigest)
	return pipe, nil
}
