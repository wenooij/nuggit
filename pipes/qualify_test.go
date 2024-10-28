package pipes

import (
	"testing"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/integrity"
)

func TestQualifyPipe(t *testing.T) {
	referencedPipeName := integrity.KeyLit(
		"foo",
		"b5cc17d3a35877ca8b76f0b2e07497039c250696",
	)
	referencedPipe := nuggit.Pipe{
		Actions: []nuggit.Action{{
			"action":   "querySelector",
			"selector": ".foo",
		}},
	}
	pipe := nuggit.Pipe{
		Actions: []nuggit.Action{{
			"action": "pipe",
			"name":   referencedPipeName.GetName(),
		}, {
			"action": "innerText",
		}},
	}

	var idx Index
	idx.Add(referencedPipeName.GetName(), referencedPipeName.GetDigest(), referencedPipe)

	qualifiedPipe, err := Qualify(&idx, pipe)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v\n", qualifiedPipe)
}
