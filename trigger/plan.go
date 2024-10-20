package trigger

import (
	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/status"
)

type planEntry struct {
	Collection *api.Collection
	Pipes      []*api.Pipe
}

type Planner struct {
	plan map[string]planEntry
}

func (p *Planner) Add(c *api.Collection, pipes []*api.Pipe) error {
	if p.plan == nil {
		p.plan = make(map[string]planEntry)
	}
	if _, found := p.plan[c.GetName()]; found {
		return nil
	}
	if err := api.ValidateCollectionPipes(c, pipes); err != nil {
		return err
	}
	p.plan[c.GetName()] = planEntry{
		Collection: c,
		Pipes:      pipes,
	}
	return nil
}

func (p *Planner) Build() *api.TriggerPlan {
	panic(status.ErrUnimplemented)
}
