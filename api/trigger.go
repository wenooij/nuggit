package api

import (
	"fmt"
	"net/url"

	"github.com/wenooij/nuggit/status"
)

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
	Collection *CollectionLite  `json:"collection,omitempty"`
	PipeArgs   map[string]*Args `json:"pipe_args,omitempty"`
}

type TriggerResponse struct {
	Storage map[string]*StorageOpLite `json:"storage,omitempty"`
}

func (a *TriggerAPI) Trigger(*TriggerRequest) (*TriggerResponse, error) {
	return nil, status.ErrUnimplemented
}

type TriggerBatchRequest struct {
	Collections []*CollectionLite `json:"collections,omitempty"`
	PipeArgs    map[string]*Args  `json:"pipe_args,omitempty"`
}

type TriggerBatchResponse struct {
	Storage []map[string]*StorageOpLite `json:"storage,omitempty"`
}

func (a *TriggerAPI) TriggerBatch(*TriggerBatchRequest) (*TriggerBatchResponse, error) {
	return nil, status.ErrUnimplemented
}

type ImplicitTriggerRequest struct {
	URL                string `json:"url,omitempty"`
	IncludeCollections bool   `json:"include_collections,omitempty"`
	IncludePipes       bool   `json:"include_pipes,omitempty"`
	IncludeStorage     bool   `json:"include_storage,omitempty"`
}

type ImplicitTriggerResponse struct {
	Collection []*CollectionLite `json:"collections,omitempty"`
	Pipes      []*PipeLite       `json:"pipes,omitempty"`
	Storage    []*StorageOpLite  `json:"storage,omitempty"`
	Actions    []*Action         `json:"client_actions,omitempty"`
}

func (a *TriggerAPI) ImplicitTrigger(req *ImplicitTriggerRequest) (*ImplicitTriggerResponse, error) {
	u, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", err, status.ErrInvalidArgument)
	}
	var triggered []*PipeLite
	if err := a.pipes.hostTriggerIndex.ScanKey(u.Hostname(), func(pipe string, err error) error {
		if err != nil {
			return err
		}
		triggered = append(triggered, newPipeLite(pipe))
		return nil
	}); err != nil {
		return nil, err
	}
	resp := &ImplicitTriggerResponse{}
	if req.IncludeCollections {
	}
	if req.IncludePipes {
		resp.Pipes = triggered
	}
	if req.IncludeStorage {
	}
	return resp, nil
}
