package trigger

import (
	"fmt"

	"github.com/wenooij/nuggit/api"
)

type Planner struct {
	g *graph

	referencedPipes map[api.NameDigest]*api.Pipe
}

type collectionEntry struct {
	Collection *api.Collection
	Pipes      []*api.Pipe
}

func (p *Planner) AddReferencedPipes(pipes []*api.Pipe) error {
	if p.referencedPipes == nil {
		p.referencedPipes = make(map[api.NameDigest]*api.Pipe, 8)
	}
	for _, pipe := range pipes {
		p.referencedPipes[pipe.GetNameDigest()] = pipe
	}
	return nil
}

func (p *Planner) Add(c *api.Collection, pipes []*api.Pipe) error {
	if p.g == nil {
		p.g = newGraph()
	}
	if err := api.ValidateCollectionPipesSubset(c, pipes); err != nil {
		return err
	}
	for i, pipe := range pipes {
		flattened, err := api.FlattenPipe(p.referencedPipes, pipe)
		if err != nil {
			return err
		}
		if err := p.g.add(pipe.GetNameDigest(), flattened.Actions); err != nil {
			return fmt.Errorf("failed to add pipe to trigger plan (#%d): %w", i, err)
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
			Action: n.action,
		})
		for _, n := range n.next {
			inputs[n] = len(steps)
		}
	}

	return &api.TriggerPlan{
		Roots:     roots,
		Exchanges: exchanges,
		Steps:     steps,
	}
}
