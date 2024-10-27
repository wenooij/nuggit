package trigger

import "github.com/wenooij/nuggit"

type Plan struct {
	// Roots is a 0-indexed list of root actions.
	Roots []int `json:"roots,omitempty"`
	// Exchanges is a 0-indexed list of exchange actions.
	Exchanges []int `json:"exchanges,omitempty"`
	// Types specifies the type number of each exchange Point.
	// The client may use this to further optimize the plan
	// or raise exceptions.
	Types []int `json:"types,omitempty"`
	// Steps contains the optimal sequence of actions needed to execute the given pipelines.
	Steps []PlanStep `json:"steps,omitempty"`
}

func (p *Plan) GetRoots() []int {
	if p == nil {
		return nil
	}
	return p.Roots
}

func (p *Plan) GetExchanges() []int {
	if p == nil {
		return nil
	}
	return p.Exchanges
}

func (p *Plan) GetSteps() []PlanStep {
	if p == nil {
		return nil
	}
	return p.Steps
}

type PlanStep struct {
	// Input is the node number representing the input to this step.
	//
	// The node number is 1-indexed, therefore equal to one greater
	// than the slice index. A value of 0 indicates the step has no
	// inputs, and that it is a root.
	Input         int `json:"input,omitempty"`
	nuggit.Action `json:",omitempty"`
}
