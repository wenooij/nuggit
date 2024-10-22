package api

import "testing"

func TestFlattenPipe(t *testing.T) {
	referencedPipeName := NameDigest{
		Name:   "foo",
		Digest: "b5cc17d3a35877ca8b76f0b2e07497039c250696",
	}
	referencedPipe := &Pipe{
		Actions: []Action{{
			Action: ActionSelector,
			Spec: &SelectorAction{
				Selector: ".foo",
			},
		}},
	}
	pipe := &Pipe{
		NameDigest: NameDigest{
			Name: "foo-text",
		},
		Actions: []Action{{
			Action: ActionPipe,
			Spec: &PipeAction{
				Pipe: referencedPipeName,
			},
		}, {
			Action: ActionField,
			Spec: &FieldAction{
				Field: "innerText",
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
