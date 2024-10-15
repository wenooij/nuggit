package api

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/wenooij/nuggit/status"
)

type TriggerAPI struct {
	api      *API
	runtimes *RuntimesAPI
	pipes    *PipesAPI
	mu       sync.Mutex
}

func (a *TriggerAPI) Init(api *API, runtimes *RuntimesAPI, pipes *PipesAPI) {
	*a = TriggerAPI{
		api:      api,
		runtimes: runtimes,
		pipes:    pipes,
	}
}

// locks excluded: api.mu, mu, pipes.mu.
func (r *TriggerAPI) run(pipe *PipeRich) error {
	p := r.pipes.pipes[pipe.ID]
	if p == nil {
		return fmt.Errorf("failed to get pipe: %w", status.ErrNotFound)
	}
	return status.ErrUnimplemented
}

type TriggerRequest struct {
	Pipe string `json:"pipe,omitempty"`
	Args *Args  `json:"args,omitempty"`
}

type TriggerResponse struct {
	Storage map[string]*StorageOpLite `json:"storage,omitempty"`
}

func (a *TriggerAPI) Trigger(*TriggerRequest) (*TriggerResponse, error) {
	return nil, status.ErrUnimplemented
}

type TriggerBatchRequest struct {
	Pipes []string `json:"pipes,omitempty"`
	Args  []*Args  `json:"args,omitempty"`
}

type TriggerBatchResponse struct {
	Storage []map[string]*StorageOpLite `json:"storage,omitempty"`
}

func (a *TriggerAPI) TriggerBatch(*TriggerBatchRequest) (*TriggerBatchResponse, error) {
	return nil, status.ErrUnimplemented
}

type ImplicitTriggerRequest struct {
	URL   string `json:"url,omitempty"`
	Quiet bool   `json:"quiet,omitempty"`
}

type ImplicitTriggerResponse struct {
	Pipes   []*PipeLite               `json:"pipes,omitempty"`
	Storage map[string]*StorageOpLite `json:"storage,omitempty"`
}

func (a *TriggerAPI) ImplicitTrigger(req *ImplicitTriggerRequest) (*ImplicitTriggerResponse, error) {
	a.api.mu.Lock()
	defer a.api.mu.Unlock()
	a.mu.Lock()
	defer a.mu.Unlock()
	a.pipes.mu.Lock()
	defer a.pipes.mu.Unlock()

	triggered := []*PipeLite{}
	for _, p := range a.pipes.alwaysTrigger {
		triggered = append(triggered, p.PipeLite)
	}
	u, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", err, status.ErrInvalidArgument)
	}
	for _, p := range a.pipes.hostTrigger[u.Hostname()] {
		triggered = append(triggered, p.PipeLite)
	}
	// TODO: Do the actual trigger here.
	if req.Quiet {
		return &ImplicitTriggerResponse{}, nil
	}
	return &ImplicitTriggerResponse{Pipes: triggered}, nil
}
