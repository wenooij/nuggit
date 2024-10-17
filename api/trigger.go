package api

import (
	"context"
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

type TriggerAPI struct {
	runtimes *RuntimesAPI
	pipes    *PipesAPI
}

func (a *TriggerAPI) Init(runtimes *RuntimesAPI, pipes *PipesAPI) {
	*a = TriggerAPI{
		runtimes: runtimes,
		pipes:    pipes,
	}
}

type TriggerRequest struct {
	Collection string `json:"collection,omitempty"`
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
	Collections []string `json:"collections,omitempty"`
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
	id, err := newUUID(func(id string) error { return status.ErrNotFound }) // FIXME
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", err, status.ErrInvalidArgument)
	}
	var triggered []*PipeLite
	var pipes []string
	if err := a.pipes.storage.ScanHostTriggered(ctx, u.Hostname(), func(pipe *PipeRich, err error) error {
		if err != nil {
			return err
		}
		triggered = append(triggered, pipe.PipeLite)
		pipes = append(pipes, pipe.UUID())
		return nil
	}); err != nil {
		return nil, err
	}

	pipesBatch, err := a.pipes.GetPipesBatch(ctx, &GetPipesBatchRequest{Pipes: pipes})
	if err != nil {
		return nil, err
	}

	collections := make(map[string]struct{})
	for _, p := range pipesBatch.Pipes {
		if p.State != nil {
			for c := range p.State.Collections {
				collections[c] = struct{}{}
			}
		}
	}

	var triggeredCollections []*CollectionLite
	for c := range collections {
		triggeredCollections = append(triggeredCollections, NewCollectionLite(c))
	}

	resp := &ImplicitTriggerResponse{Trigger: NewTriggerLite(id)}
	if req.IncludeCollections {
		resp.Collection = triggeredCollections
	}
	if req.IncludePipes {
		resp.Pipes = triggered
	}
	return resp, nil
}
