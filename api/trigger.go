package api

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/url"
	"slices"
	"time"

	"github.com/wenooij/nuggit/status"
)

const triggersBaseURI = "/api/triggers"

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

func (p *TriggerPlan) GetRoots() []int {
	if p == nil {
		return nil
	}
	return p.Roots
}

func (p *TriggerPlan) GetExchanges() []int {
	if p == nil {
		return nil
	}
	return p.Exchanges
}

func (p *TriggerPlan) GetSteps() []TriggerPlanStep {
	if p == nil {
		return nil
	}
	return p.Steps
}

type TriggerPlanStep struct {
	// Input is the node number representing the input to this step.
	//
	// The node number is 1-indexed, therefore equal to one greater
	// than the slice index. A value of 0 indicates the step has no
	// inputs, and that it is a root.
	Input  int `json:"input,omitempty"`
	Action `json:",omitempty"`
}

type TriggerRecord struct {
	*Trigger     `json:"trigger,omitempty"`
	*TriggerPlan `json:"plan,omitempty"`
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

type ExchangeResult struct {
	Pipe   NameDigest      `json:"pipe,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
}

func (r *ExchangeResult) GetPipe() NameDigest {
	if r == nil {
		return NameDigest{}
	}
	return r.Pipe
}

func (r *ExchangeResult) GetResult() json.RawMessage {
	if r == nil {
		return nil
	}
	return r.Result
}

type TriggerResult struct {
	Trigger        string `json:"trigger,omitempty"`
	ExchangeResult `json:","`
}

type TriggerAPI struct {
	store       TriggerStore
	results     ResultStore
	collections *CollectionsAPI
	pipes       *PipesAPI
	newPlanner  func() TriggerPlanner
}

func (a *TriggerAPI) Init(store TriggerStore, newPlanner func() TriggerPlanner, results ResultStore, collections *CollectionsAPI, pipes *PipesAPI) {
	*a = TriggerAPI{
		store:       store,
		results:     results,
		collections: collections,
		pipes:       pipes,
		newPlanner:  newPlanner,
	}
}

type GetTriggerRequest struct {
	Trigger string `json:"trigger,omitempty"`
}

type GetTriggerResponse struct {
	*TriggerRecord `json:",omitempty"`
}

func (t *TriggerAPI) GetTrigger(ctx context.Context, req *GetTriggerRequest) (*GetTriggerResponse, error) {
	trigger, err := t.store.Load(ctx, req.Trigger)
	if err != nil {
		return nil, err
	}
	return &GetTriggerResponse{TriggerRecord: trigger}, nil
}

type CreateTriggerPlanRequest struct {
	*Trigger            `json:"trigger,omitempty"`
	IncludeCollections  []string `json:"include_collections,omitempty"`
	ExcludeCollections  []string `json:"exclude_collections,omitempty"`
	PopulateCollections bool     `json:"populate_collections,omitempty"`
	PopulatePipes       bool     `json:"populate_pipes,omitempty"`
}

type CreateTriggerPlanResponse struct {
	Trigger *Ref         `json:"trigger,omitempty"`
	Plan    *TriggerPlan `json:"plan,omitempty"`
}

func (a *TriggerAPI) CreateTriggerPlan(ctx context.Context, req *CreateTriggerPlanRequest) (*CreateTriggerPlanResponse, error) {
	if err := provided("trigger", "is", req.Trigger); err != nil {
		return nil, err
	}
	u, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", err, status.ErrInvalidArgument)
	}

	pipes := make(map[NameDigest]*Pipe)
	collections := make(map[NameDigest]*Collection)
	collectionPipes := make(map[NameDigest][]*Pipe)

	for e, err := range a.collections.store.ScanTriggered(ctx, u) {
		if err != nil {
			return nil, err
		}
		pipes[e.Pipe.GetNameDigest()] = e.Pipe
		collectionName := e.Collection.GetNameDigest()
		collections[collectionName] = e.Collection
		collectionPipes[collectionName] = append(collectionPipes[collectionName], e.Pipe)
	}

	tp := a.newPlanner()

	// Add referenced pipes to Plan.
	// This is required for the FlattenPipes calls later on.
	for _, p := range pipes {
		// TODO: Maybe this query can be made batch?
		for rp, err := range a.pipes.store.ScanPipeReferences(ctx, p.GetNameDigest()) {
			if err != nil {
				return nil, err
			}
			if err := tp.AddReferencedPipes([]*Pipe{rp}); err != nil {
				return nil, err
			}
		}
	}

	// Add pipes to Plan.
	for collectionName, pipes := range collectionPipes {
		c := collections[collectionName]
		if err := tp.Add(c, pipes); err != nil {
			return nil, err
		}
	}

	plan := tp.Build()
	if len(plan.GetSteps()) == 0 {
		// Plan is a no-op.
		// Don't store the trigger and send empty response.
		return &CreateTriggerPlanResponse{}, nil
	}

	// Store the trigger and return it since is isn't a no-op.
	trigger, err := a.store.Store(ctx, &TriggerRecord{
		Trigger:     req.Trigger,
		TriggerPlan: plan,
	})
	if err != nil {
		return nil, err
	}

	// Store triggered collections.
	if err := a.store.StoreTriggerCollections(ctx, trigger, slices.Collect(maps.Keys(collections))); err != nil {
		return nil, err
	}

	triggerRef := newRef(triggersBaseURI, trigger)
	return &CreateTriggerPlanResponse{
		Trigger: &triggerRef,
		Plan:    plan,
	}, nil
}

type ExchangeResultsRequest struct {
	Trigger string           `json:"trigger,omitempty"`
	Results []ExchangeResult `json:"results,omitempty"`
}

type ExchangeResultsResponse struct{}

func (a *TriggerAPI) ExchangeResults(ctx context.Context, req *ExchangeResultsRequest) (*ExchangeResultsResponse, error) {
	if err := provided("trigger", "is", req.Trigger); err != nil {
		return nil, err
	}
	collections := make(map[NameDigest]*Collection)
	for c, err := range a.store.ScanTriggerCollections(ctx, req.Trigger) {
		if err != nil {
			return nil, err
		}
		collections[c.NameDigest] = c
	}
	collectionPipes := make(map[NameDigest][]*Pipe)
	for cp, err := range a.collections.store.ScanCollectionPipes(ctx) {
		if err != nil {
			return nil, err
		}
		name := cp.Collection.GetNameDigest()
		collectionPipes[name] = append(collectionPipes[name], cp.Pipe)
	}
	pipeResults := make(map[NameDigest]ExchangeResult)
	for _, r := range req.Results {
		pipeResults[r.Pipe] = r
	}
	for name, c := range collections {
		pipes := collectionPipes[name]
		results := []ExchangeResult{}
		for _, p := range c.GetPipes() {
			results = append(results, pipeResults[p])
		}
		if err := a.results.InsertRow(ctx, c, pipes, results); err != nil {
			return nil, err
		}
	}
	return &ExchangeResultsResponse{}, nil
}

type CommitTriggerRequest struct {
	Trigger string `json:"trigger,omitempty"`
}

type CommitTriggerResponse struct{}

func (a *TriggerAPI) CommitTrigger(ctx context.Context, req *CommitTriggerRequest) (*CommitTriggerResponse, error) {
	if err := a.store.Commit(ctx, req.Trigger); err != nil {
		return nil, err
	}
	return &CommitTriggerResponse{}, nil
}
