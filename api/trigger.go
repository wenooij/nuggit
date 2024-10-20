package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/wenooij/nuggit/status"
)

func NewTriggerRef(id string) *Ref {
	return newRef("/api/triggers/", id)
}

type Trigger struct {
	Implicit  bool      `json:"implicit,omitempty"`
	URL       string    `json:"url,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

func (t *Trigger) GetImplicit() bool {
	if t == nil {
		return false
	}
	return t.Implicit
}

func (t *Trigger) GetURL() string {
	if t == nil {
		return ""
	}
	return t.URL
}

func (t *Trigger) GetTimestamp() time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.Timestamp
}

type TriggerPlan struct {
	// Roots is a 0-indexed list of root actions.
	Roots []int `json:"roots,omitempty"`
	// Exchanges is a 0-indexed list of exchange actions.
	Exchanges []int `json:"exchanges,omitempty"`
	// Steps contains the optimal sequence of actions needed to execute the given pipelines.
	Steps []TriggerPlanStep `json:"steps,omitempty"`
}

type TriggerPlanStep struct {
	// Input is the node number representing the input to this step.
	//
	// The node number is 1-indexed, therefore equal to one greater
	// than the slice index. A value of 0 indicates the step has no
	// inputs, and that it is a root.
	Input   int `json:"input,omitempty"`
	*Action `json:",omitempty"`
}

type TriggerRecord struct {
	*Trigger     `json:",omitempty"`
	*TriggerPlan `json:",omitempty"`
}

func (t *TriggerRecord) GetTrigger() *Trigger {
	if t == nil {
		return nil
	}
	return t.Trigger
}

func (t *TriggerRecord) GetPlan() *TriggerPlan {
	if t == nil {
		return nil
	}
	return t.TriggerPlan
}

type TriggerResult struct {
	Trigger string          `json:"trigger,omitempty"`
	Pipe    string          `json:"pipe,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

type TriggerAPI struct {
	store       TriggerStore
	results     StoreInterface[*TriggerResult]
	collections *CollectionsAPI
	pipes       *PipesAPI
}

func (a *TriggerAPI) Init(store TriggerStore, results StoreInterface[*TriggerResult], collections *CollectionsAPI, pipes *PipesAPI) {
	*a = TriggerAPI{
		store:       store,
		collections: collections,
		pipes:       pipes,
	}
}

type CreateTriggerPlanRequest struct {
	*Trigger            `json:"trigger,omitempty"`
	IncludeCollections  []string `json:"include_collections,omitempty"`
	ExcludeCollections  []string `json:"exclude_collections,omitempty"`
	PopulateCollections bool     `json:"populate_collections,omitempty"`
	PopulatePipes       bool     `json:"populate_pipes,omitempty"`
}

type CreateTriggerPlanResponse struct {
	Trigger    *Ref         `json:"trigger,omitempty"`
	Collection []*Ref       `json:"collections,omitempty"`
	Pipes      []string     `json:"pipes,omitempty"`
	Plan       *TriggerPlan `json:"plan,omitempty"`
}

func (a *TriggerAPI) CreateTriggerPlan(ctx context.Context, req *CreateTriggerPlanRequest) (*CreateTriggerPlanResponse, error) {
	if err := provided("trigger", "is", req.Trigger); err != nil {
		return nil, err
	}
	u, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", err, status.ErrInvalidArgument)
	}
	var uniqueCollections []*Ref
	uniquePipes := make(map[string]struct{})
	if err := a.collections.store.ScanTriggered(ctx, u, func(id string, collection *Collection, err error) error {
		if err != nil {
			return err
		}
		uniqueCollections = append(uniqueCollections, NewCollectionRef(id))
		for _, p := range collection.GetPipes() {
			uniquePipes[p] = struct{}{}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	pipes := make([]string, 0, len(uniquePipes))
	uniquePipeRefs := make([]string, 0, len(uniquePipes))
	for p := range uniquePipes {
		pipes = append(pipes, p)
		uniquePipeRefs = append(uniquePipeRefs, p)
	}
	// TODO: Use a Scan here for better performance.
	pipesBatch, err := a.pipes.GetPipesBatch(ctx, &GetPipesBatchRequest{IDs: pipes})
	if err != nil {
		return nil, err
	}

	if len(pipesBatch.Missing) != 0 {
		return nil, fmt.Errorf("required pipes referenced by collections are missing: %w", status.ErrDataLoss)
	}

	plan := &TriggerPlan{}
	for i, p := range pipesBatch.Pipes {
		actions := p.GetActions()
		if len(actions) == 0 {
			continue
		}
		plan.Roots = append(plan.Roots, len(plan.Steps))
		for i, a := range actions {
			copyAction := a
			step := TriggerPlanStep{
				Action: &copyAction,
			}
			if i > 0 {
				step.Input = len(plan.Steps) - 1
			}
			plan.Steps = append(plan.Steps, step)
		}
		plan.Exchanges = append(plan.Exchanges, len(plan.Steps))
		exchangeAction := &Action{
			Action: ActionExchange,
			Spec: map[string]any{
				"pipe": pipes[i],
			}, // = Exchange{Pipe: pipes[i]}.
		}
		exchangeStep := TriggerPlanStep{
			Input:  len(plan.Steps) - 1,
			Action: exchangeAction,
		}
		plan.Steps = append(plan.Steps, exchangeStep)
	}

	if len(plan.Steps) == 0 {
		// Return early without storing the trigger.
		// We'll return 204 No Content to indicate
		// we didn't do anything.
		return &CreateTriggerPlanResponse{}, nil
	}

	id, err := a.store.Store(ctx, &TriggerRecord{
		Trigger:     req.Trigger,
		TriggerPlan: plan,
	})
	if err != nil {
		return nil, err
	}

	resp := &CreateTriggerPlanResponse{
		Trigger: NewTriggerRef(id),
		Plan:    plan,
	}
	if req.PopulateCollections {
		resp.Collection = uniqueCollections
	}
	if req.PopulatePipes {
		resp.Pipes = uniquePipeRefs
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

type CommitCollectionRequest struct {
	Trigger    string `json:"trigger,omitempty"`
	Collection string `json:"collection,omitempty"`
}

type CommitCollectionResponse struct{}

func (a *TriggerAPI) CommitCollection(ctx context.Context, req *CommitCollectionRequest) (*CommitCollectionResponse, error) {
	return nil, status.ErrUnimplemented
}

type CommitTriggerRequest struct {
	Trigger string `json:"trigger,omitempty"`
}

type CommitTriggerResponse struct{}

func (a *TriggerAPI) CommitTrigger(ctx context.Context, req *CommitTriggerRequest) (*CommitTriggerResponse, error) {
	return nil, status.ErrUnimplemented
}
