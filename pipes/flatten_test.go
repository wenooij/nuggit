package pipes

import (
	"testing"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/integrity"
)

func TestFlattenPipe(t *testing.T) {
	referencedPipeName := integrity.NameDigest{
		Name:   "foo",
		Digest: "b5cc17d3a35877ca8b76f0b2e07497039c250696",
	}
	referencedPipe := nuggit.Pipe{
		Actions: []nuggit.Action{{
			"action":   "querySelector",
			"selector": ".foo",
		}},
	}
	pipe := nuggit.Pipe{
		Actions: []nuggit.Action{{
			"action": "pipe",
			"name":   referencedPipeName.Name,
			"digest": referencedPipeName.Digest,
		}, {
			"action": "innerText",
		}},
	}

	referencedPipes := map[integrity.NameDigest]nuggit.Pipe{referencedPipeName: referencedPipe}

	flattenedPipe, err := Flatten(referencedPipes, pipe)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v\n", flattenedPipe)
}
