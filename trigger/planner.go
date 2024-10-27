package trigger

import (
	"fmt"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/integrity"
	pipeutil "github.com/wenooij/nuggit/pipes"
)

type Planner struct {
	g *graph

	referencedPipes map[integrity.NameDigest]nuggit.Pipe
}

func (p *Planner) AddReferencedPipe(name, digest string, pipe nuggit.Pipe) {
	if p.referencedPipes == nil {
		p.referencedPipes = make(map[integrity.NameDigest]nuggit.Pipe, 8)
	}
	p.referencedPipes[integrity.KeyLit(name, digest)] = pipe
}

func (p *Planner) AddPipe(name, digest string, pipe nuggit.Pipe) error {
	if p.g == nil {
		p.g = newGraph()
	}
	flattened, err := pipeutil.Flatten(p.referencedPipes, pipe)
	if err != nil {
		return err
	}
	if err := p.g.add(integrity.KeyLit(name, digest), pipe, flattened.Actions); err != nil {
		return fmt.Errorf("failed to add pipe to trigger plan: %w", err)
	}
	return nil
}

func (p *Planner) Build() *Plan {
	n := p.g.Len()
	roots := make([]int, 0, n)
	exchanges := make([]int, 0, n)
	steps := make([]PlanStep, 0, 64)
	inputs := make(map[*graphNode]int, 64)

	for n := range p.g.consistentTopoIter {
		i := len(steps)
		input := inputs[n]
		if input == 0 {
			roots = append(roots, i)
		}
		if len(n.next) == 0 && n.action.GetAction() == "exchange" {
			exchanges = append(exchanges, i)
		}
		steps = append(steps, PlanStep{
			Input:  inputs[n],
			Action: n.action,
		})
		for _, n := range n.next {
			inputs[n] = len(steps)
		}
	}

	return &Plan{
		Roots:     roots,
		Exchanges: exchanges,
		Steps:     steps,
	}
}
