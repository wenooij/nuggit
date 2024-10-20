package trigger

import (
	"github.com/wenooij/nuggit/api"
)

type Planner struct {
	g *graph
}

type collectionEntry struct {
	Collection *api.Collection
	Pipes      []*api.Pipe
}

func (p *Planner) Add(c *api.Collection, pipes []*api.Pipe) error {
	if p.g == nil {
		p.g = newGraph()
	}
	if err := api.ValidateCollectionPipes(c, pipes); err != nil {
		return err
	}
	for _, pipe := range pipes {
		if err := p.g.add(pipe); err != nil {
			return err
		}
	}
	return nil
}

func (p *Planner) Build() *api.TriggerPlan {
	n := p.g.Len()
	roots := make([]int, 0, n)
	exchanges := make([]int, 0, n)
	steps := make([]api.TriggerPlanStep, 0, 64)
	inputs := make(map[*graphNode]int, 64)

	for n := range p.g.consistentTopoIter {
		i := len(steps)
		input := inputs[n]
		if input == 0 {
			roots = append(roots, i)
		}
		if len(n.next) == 0 && n.action.GetAction() == api.ActionExchange {
			exchanges = append(exchanges, i)
		}
		steps = append(steps, api.TriggerPlanStep{
			Input:  inputs[n],
			Action: &n.action,
		})
		for _, n := range n.next {
			inputs[n] = 1 + i
		}
	}

	return &api.TriggerPlan{
		Roots:     roots,
		Exchanges: exchanges,
		Steps:     steps,
	}
}