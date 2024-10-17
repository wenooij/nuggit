package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/wenooij/nuggit/status"
)

type TriggerLite struct {
	*Ref
}

func NewTriggerLite(id string) *TriggerLite {
	return &TriggerLite{newRef("/api/triggers/%s", id)}
}

type TriggerBase struct {
	Result json.RawMessage `json:"result,omitempty"`
}

type Trigger struct {
	*TriggerLite `json:",omitempty"`
	*TriggerBase `json:",omitempty"`
	Pipe         *PipeLite `json:"pipe,omitempty"`
}

type TriggerAPI struct {
	store       StoreInterface[*Trigger]
	collections *CollectionsAPI
	pipes       *PipesAPI
}

func (a *TriggerAPI) Init(store StoreInterface[*Trigger], collections *CollectionsAPI, pipes *PipesAPI) {
	*a = TriggerAPI{
		store:       store,
		collections: collections,
		pipes:       pipes,
	}
}

type TriggerRequest struct {
	Collection    string `json:"collection,omitempty"`
	PopulatePipes bool   `json:"populate_pipes,omitempty"`
}

type TriggerResponse struct {
	Trigger *TriggerLite `json:"trigger,omitempty"`
	Pipes   []*PipeLite  `json:"pipes,omitempty"`
	Actions []*Action    `json:"actions,omitempty"`
}

func (a *TriggerAPI) Trigger(context.Context, *TriggerRequest) (*TriggerResponse, error) {
	return nil, status.ErrUnimplemented
}

type TriggerBatchRequest struct {
	Collections   []string `json:"collections,omitempty"`
	PopulatePipes bool     `json:"populate_pipes,omitempty"`
}

type TriggerBatchResponse struct {
	Trigger *TriggerLite `json:"triggers,omitempty"`
	Pipes   []*PipeLite  `json:"pipes,omitempty"`
	Actions []*Action    `json:"actions,omitempty"`
}

func (a *TriggerAPI) TriggerBatch(context.Context, *TriggerBatchRequest) (*TriggerBatchResponse, error) {
	return nil, status.ErrUnimplemented
}

type ImplicitTriggerRequest struct {
	URL                string `json:"url,omitempty"`
	IncludeCollections bool   `json:"include_collections,omitempty"`
	IncludePipes       bool   `json:"include_pipes,omitempty"`
}

type ImplicitTriggerResponse struct {
	Trigger    *TriggerLite      `json:"trigger,omitempty"`
	Collection []*CollectionLite `json:"collections,omitempty"`
	Pipes      []*PipeLite       `json:"pipes,omitempty"`
	Actions    []*Action         `json:"actions,omitempty"`
}

func (a *TriggerAPI) ImplicitTrigger(ctx context.Context, req *ImplicitTriggerRequest) (*ImplicitTriggerResponse, error) {
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
		for p := range collection.GetState().GetPipes() {
			uniquePipes[p] = struct{}{}
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

	actions := make([]*Action, 0)
	for _, p := range pipesBatch.Pipes {
		for _, n := range p.GetBase().GetSequence() {
			actions = append(actions, n.GetAction())
		}
	}

	resp := &ImplicitTriggerResponse{
		Trigger: NewTriggerLite(id),
		Actions: actions,
	}
	if req.IncludeCollections {
		resp.Collection = uniqueCollections
	}
	if req.IncludePipes {
		resp.Pipes = uniquePipeLites
	}
	return resp, nil
}
