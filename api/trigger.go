package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/wenooij/nuggit/status"
)

type TriggerLite struct {
	*Ref
}

func NewTriggerLite(id string) *TriggerLite {
	return &TriggerLite{newRef("/api/triggers/%s", id)}
}

type TriggerBase struct {
	Implicit  bool      `json:"implicit,omitempty"`
	URL       string    `json:"url,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

type Trigger struct {
	*TriggerLite `json:",omitempty"`
	*TriggerBase `json:",omitempty"`
}

type TriggerPlan struct {
	Roots     []int             `json:"roots,omitempty"`
	Exchanges []int             `json:"exchanges,omitempty"`
	Steps     []TriggerPlanStep `json:"steps,omitempty"`
}

type TriggerPlanStep struct {
	Input   int `json:"input,omitempty"`
	*Action `json:",omitempty"`
}

type TriggerResult struct {
	*TriggerLite `json:",omitempty"`
	Pipe         *PipeLite       `json:",omitempty"`
	Result       json.RawMessage `json:"result,omitempty"`
}

type TriggerAPI struct {
	store       StoreInterface[*Trigger]
	results     StoreInterface[*TriggerResult]
	collections *CollectionsAPI
	pipes       *PipesAPI
}

func (a *TriggerAPI) Init(store StoreInterface[*Trigger], results StoreInterface[*TriggerResult], collections *CollectionsAPI, pipes *PipesAPI) {
	*a = TriggerAPI{
		store:       store,
		collections: collections,
		pipes:       pipes,
	}
}

type GetTriggerPlanRequest struct {
	*TriggerBase        `json:",omitempty"`
	IncludeCollections  []string `json:"include_collections,omitempty"`
	ExcludeCollections  []string `json:"exclude_collections,omitempty"`
	PopulateCollections bool     `json:"populate_collections,omitempty"`
	PopulatePipes       bool     `json:"populate_pipes,omitempty"`
}

type GetTriggerPlanResponse struct {
	Trigger    *TriggerLite      `json:"trigger,omitempty"`
	Collection []*CollectionLite `json:"collections,omitempty"`
	Pipes      []*PipeLite       `json:"pipes,omitempty"`
	Plan       *TriggerPlan      `json:"plan,omitempty"`
}

func (a *TriggerAPI) GetTriggerPlan(ctx context.Context, req *GetTriggerPlanRequest) (*GetTriggerPlanResponse, error) {
	id, err := newUUID(ctx, a.store.Exists)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", err, status.ErrInvalidArgument)
	}
	var uniqueCollections []*CollectionLite
	var uniquePipes map[string]struct{}
	if err := a.collections.store.ScanTriggered(ctx, u, func(collection *Collection, err error) error {
		if err != nil {
			return err
		}
		uniqueCollections = append(uniqueCollections, collection.GetLite())
		for _, p := range collection.GetBase().GetPipes() {
			uniquePipes[p.GetRef().UUID()] = struct{}{}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	pipes := make([]string, 0, len(uniquePipes))
	uniquePipeLites := make([]*PipeLite, 0, len(uniquePipes))
	for p := range uniquePipes {
		pipes = append(pipes, p)
		uniquePipeLites = append(uniquePipeLites, NewPipeLite(p))
	}
	// TODO: Use a Scan here for better performance.
	pipesBatch, err := a.pipes.GetPipesBatch(ctx, &GetPipesBatchRequest{Pipes: pipes})
	if err != nil {
		return nil, err
	}

	if len(pipesBatch.Missing) != 0 {
		return nil, fmt.Errorf("required pipes referenced by collections are missing: %w", status.ErrDataLoss)
	}

	plan := &TriggerPlan{}
	for _, p := range pipesBatch.Pipes {
		actions := p.GetBase().GetActions()
		if len(actions) == 0 {
			continue
		}
		plan.Roots = append(plan.Roots, len(plan.Steps))
		for i, a := range actions {
			step := TriggerPlanStep{
				Action: a,
			}
			if i > 0 {
				step.Input = len(plan.Steps) - 1
			}
			plan.Steps = append(plan.Steps, step)
		}
		plan.Exchanges = append(plan.Exchanges, len(plan.Steps))
		exchangeSpec := &ExchangeAction{Pipe: p.GetLite().GetRef().UUID()}
		exchangeSpecBytes, err := json.Marshal(exchangeSpec)
		if err != nil {
			return nil, err
		}
		exchangeAction := &Action{
			Action: ActionExchange,
			Spec:   exchangeSpecBytes,
		}
		exchangeStep := TriggerPlanStep{
			Input:  len(plan.Steps) - 1,
			Action: exchangeAction,
		}
		plan.Steps = append(plan.Steps, exchangeStep)
	}

	resp := &GetTriggerPlanResponse{
		Trigger: NewTriggerLite(id),
		Plan:    plan,
	}
	if req.PopulateCollections {
		resp.Collection = uniqueCollections
	}
	if req.PopulatePipes {
		resp.Pipes = uniquePipeLites
	}
	return resp, nil
}

type ExchangeResultRequest struct {
	Trigger string          `json:"trigger,omitempty"`
	Pipe    string          `json:"pipe,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

type ExchangeResultResponse struct{}

func (a *TriggerAPI) ExchangeResult(ctx context.Context, req *ExchangeResultRequest) (*ExchangeResultResponse, error) {
	return nil, status.ErrUnimplemented
}
