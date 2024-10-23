package api

import "testing"

func TestFlattenPipe(t *testing.T) {
	referencedPipeName := NameDigest{
		Name:   "foo",
		Digest: "b5cc17d3a35877ca8b76f0b2e07497039c250696",
	}
	referencedPipe := &Pipe{
		Actions: []Action{{
			Action: "querySelector",
			Args: map[string]string{
				"selector": ".foo",
			},
		}},
	}
	pipe := &Pipe{
		NameDigest: NameDigest{
			Name: "foo-text",
		},
		Actions: []Action{{
			Action: "pipe",
			Args: map[string]string{
				"name":   referencedPipeName.Name,
				"digest": referencedPipeName.Digest,
			},
		}, {
			Action: "method",
			Args: map[string]string{
				"method": "innerText",
			},
		}},
	}

	referencedPipes := map[NameDigest]*Pipe{referencedPipeName: referencedPipe}

	flattenedPipe, err := FlattenPipe(referencedPipes, pipe)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%#v\n", flattenedPipe)
}
