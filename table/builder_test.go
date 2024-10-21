package table

import (
	"testing"

	"github.com/wenooij/nuggit/api"
)

func TestBuilder(t *testing.T) {
	c := api.Collection{
		NameDigest: api.NameDigest{
			Name:   "foo",
			Digest: "8974e2039150e9c0492bdb1a46359963d2a4c74a",
		},
		Pipes: []api.NameDigest{{
			Name:   "foo1",
			Digest: "bc4537ecb89d71648e6f2e2b4c8b43be46d24589",
		}, {
			Name:   "foo2",
			Digest: "c8965d7dc715a6f46350ce5ce5fe3d129c7995af",
		}, {
			Name:   "foo3",
			Digest: "1dac61e57cd2b5616d5f18d0bd9c955bb878282a",
		}},
	}
	pipes := []*api.Pipe{{
		NameDigest: api.NameDigest{
			Name:   "foo1",
			Digest: "bc4537ecb89d71648e6f2e2b4c8b43be46d24589",
		},
	}, {
		NameDigest: api.NameDigest{
			Name:   "foo2",
			Digest: "c8965d7dc715a6f46350ce5ce5fe3d129c7995af",
		},
	}, {
		NameDigest: api.NameDigest{
			Name:   "foo3",
			Digest: "1dac61e57cd2b5616d5f18d0bd9c955bb878282a",
		},
	}}

	var b Builder
	b.Reset(&c)
	if err := b.Add(pipes...); err != nil {
		t.Fatal(err)
	}
	expr, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(expr)
}
