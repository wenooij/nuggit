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
		collections: collections,
		pipes:       pipes,
		newPlanner:  newPlanner,
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
	Trigger     *Ref         `json:"trigger,omitempty"`
	Collections []Ref        `json:"collections,omitempty"`
	Pipes       []Ref        `json:"pipes,omitempty"`
	Plan        *TriggerPlan `json:"plan,omitempty"`
}

func (a *TriggerAPI) CreateTriggerPlan(ctx context.Context, req *CreateTriggerPlanRequest) (*CreateTriggerPlanResponse, error) {
	if err := provided("trigger", "is", req.Trigger); err != nil {
		return nil, err
	}
	u, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", err, status.ErrInvalidArgument)
	}

	tp := a.newPlanner()

	collectionPipes := make(map[NameDigest]struct{})
	uniqueCollections := make([]Ref, 0)

	for e, err := range a.collections.store.ScanTriggered(ctx, u) {
		if err != nil {
			return nil, err
		}
		if err := tp.Add(e.Collection, []*Pipe{e.Pipe}); err != nil {
			return nil, err
		}
		uniqueCollections = append(uniqueCollections, newNamedRef(collectionsBaseURI, e.Collection.NameDigest))
		collectionPipes[e.Pipe.GetNameDigest()] = struct{}{}
	}

	slices.SortFunc(uniqueCollections, compareRef)
	uniquePipes := slices.SortedFunc(maps.Keys(collectionPipes), compareNameDigest)
	uniquePipeRefs := make([]Ref, 0, len(uniquePipes))
	for _, p := range uniquePipes {
		uniquePipeRefs = append(uniquePipeRefs, newNamedRef(pipesBaseURI, p))
	}

	plan := tp.Build()

	resp := &CreateTriggerPlanResponse{}

	if len(plan.GetSteps()) != 0 {
		// Store the trigger since is isn't a no-op.
		id, err := a.store.Store(ctx, &TriggerRecord{
			Trigger:     req.Trigger,
			TriggerPlan: plan,
		})
		if err != nil {
			return nil, err
		}
		ref := newRef(triggersBaseURI, id)
		resp.Trigger = &ref
		resp.Plan = plan
	}
	if req.PopulateCollections {
		resp.Collections = uniqueCollections
	}
	if req.PopulatePipes {
		resp.Pipes = uniquePipeRefs
	}
	return resp, nil
}

type ExchangeResultsRequest struct {
	Trigger string           `json:"trigger,omitempty"`
	Results []ExchangeResult `json:"results,omitempty"`
}

type ExchangeResultsResponse struct{}

func (a *TriggerAPI) ExchangeResults(ctx context.Context, req *ExchangeResultsRequest) (*ExchangeResultsResponse, error) {
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
