package pipes

import (
	"testing"

	"github.com/wenooij/nuggit"
)

func TestQualified(t *testing.T) {
	foo := nuggit.Pipe{
		Actions: []nuggit.Action{{
			"action":   "querySelector",
			"selector": ".foo",
		}},
	}
	bar := nuggit.Pipe{
		Actions: []nuggit.Action{{
			"action": "pipe",
			"name":   "foo",
		}, {
			"action": "innerText",
		}},
	}

	var idx Index
	idx.Add("foo", "", foo)
	idx.Add("bar", "", bar)

	qualified, err := idx.Qualified()
	if err != nil {
		t.Fatal(err)
	}

	for nd := range qualified.Topo() {
		pipe, ok := qualified.Get(nd.GetName(), nd.GetDigest())
		t.Log(nd, pipe, ok)
	}
}
