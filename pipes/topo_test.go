package pipes

import (
	"testing"

	"github.com/wenooij/nuggit"
)

func TestIndexTopo(t *testing.T) {
	var i Index

	i.Add("doo", "", nuggit.Pipe{Actions: []nuggit.Action{
		{"action": "pipe", "name": "baz"},
		{"action": "pipe", "name": "foo"},
	}})
	i.Add("baz", "", nuggit.Pipe{Actions: []nuggit.Action{{"action": "pipe", "name": "bar"}}})
	i.Add("bar", "", nuggit.Pipe{Actions: []nuggit.Action{{"action": "pipe", "name": "foo"}}})
	i.Add("foo", "", nuggit.Pipe{})

	for nd, err := range i.Topo() {
		t.Log(nd, err)
	}
}
