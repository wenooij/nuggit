package trigger

import (
	"testing"

	"github.com/wenooij/nuggit"
)

func TestTriggerPlanner(t *testing.T) {
	pipe := nuggit.Pipe{
		Actions: []nuggit.Action{{
			"action": "a1",
		}, {
			"action": "a2",
		}, {
			"action": "a3",
		}},
	}

	var p Planner
	if err := p.AddPipe("foo", "123", pipe); err != nil {
		t.Fatal(err)
	}
	if err := p.AddPipe("foo", "456", pipe); err != nil {
		t.Fatal(err)
	}
	plan := p.Build()
	t.Log(plan)
}
