package table

import (
	"testing"

	"github.com/google/uuid"
	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/integrity"
)

func TestViewBuilder(t *testing.T) {
	v := &api.View{
		Alias: "foo",
		Columns: []api.ViewColumn{{
			Pipe: &api.Pipe{
				NameDigest: integrity.NameDigest{
					Name:   "foo1",
					Digest: "bc4537ecb89d71648e6f2e2b4c8b43be46d24589",
				},
			},
		}, {
			Pipe: &api.Pipe{
				NameDigest: integrity.NameDigest{
					Name:   "foo2",
					Digest: "c8965d7dc715a6f46350ce5ce5fe3d129c7995af",
				},
			},
		}, {
			Pipe: &api.Pipe{
				NameDigest: integrity.NameDigest{
					Name:   "foo3",
					Digest: "1dac61e57cd2b5616d5f18d0bd9c955bb878282a",
				},
			},
		}},
	}

	var b ViewBuilder
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	if err := b.SetView(id.String(), v.Alias); err != nil {
		t.Fatal(err)
	}
	for _, col := range v.Columns {
		if err := b.AddViewColumn(col); err != nil {
			t.Fatal(err)
		}
	}
	expr, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(expr)
}
