package pipes

import (
	"testing"

	"github.com/wenooij/nuggit/api"
)

func TestFlattenPipe(t *testing.T) {
	referencedPipeName := api.NameDigest{
		Name:   "foo",
		Digest: "b5cc17d3a35877ca8b76f0b2e07497039c250696",
	}
	referencedPipe := &api.Pipe{
		Actions: []api.Action{{
			"action":   "querySelector",
			"selector": ".foo",
		}},
	}
	pipe := &api.Pipe{
		NameDigest: api.NameDigest{
			Name: "foo-text",
		},
		Actions: []api.Action{
			api.MakePipeAction(referencedPipeName),
			{
				"action": "innerText",
			}},
	}

	referencedPipes := map[api.NameDigest]*api.Pipe{referencedPipeName: referencedPipe}

	flattenedPipe, err := Flatten(referencedPipes, pipe)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v\n", flattenedPipe)
}
