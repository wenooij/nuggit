package api

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/wenooij/nuggit/status"
)

type RuntimeLite struct {
	*Ref `json:",omitempty"`
}

type RuntimeBase struct {
	Name             string        `json:"name,omitempty"`
	SupportedActions []*ActionLite `json:"supported_actions,omitempty"`
}

type Runtime struct {
	*RuntimeLite `json:",omitempty"`
	*RuntimeBase `json:",omitempty"`
}

type RuntimesAPI struct {
	api   *API
	pipes *PipesAPI
	rc    map[string]*Runtime
	mu    sync.RWMutex
}

func (a *RuntimesAPI) Init(api *API, pipes *PipesAPI) error {
	// Add builtin pipe.
	_, err := pipes.CreatePipe(&CreatePipeRequest{
		Pipe: &PipeBaseRich{
			Sequence: []*NodeBase{{
				Action: "document",
			}},
		},
	})
	if err != nil {
		return err
	}
	*a = RuntimesAPI{
		api:   api,
		pipes: pipes,
		rc:    make(map[string]*Runtime),
	}
	id, _ := newUUID(func(id string) bool { return true })
	a.createRuntime(&Runtime{
		RuntimeLite: &RuntimeLite{
			Ref: &Ref{
				ID:  id,
				URI: fmt.Sprintf("/api/runtimes/%s", id),
			},
		},
		RuntimeBase: &RuntimeBase{
			Name: "default",
		},
	}) // Always returns true.
	return nil
}

// locks excluded: api.mu, mu, pipes.mu.
func (r *RuntimesAPI) run(pipe *PipeRich) error {
	p := r.pipes.pipes[pipe.ID]
	if p == nil {
		return fmt.Errorf("failed to get pipe: %w", status.ErrNotFound)
	}
	return status.ErrUnimplemented
}

// locks excluded: mu.
func (r *RuntimesAPI) createRuntime(rt *Runtime) bool {
	if r.rc[rt.ID] != nil {
		return false
	}
	r.rc[rt.ID] = rt
	return true
}

type TriggerRequest struct {
	Pipe string `json:"pipe,omitempty"`
	Args *Args  `json:"args,omitempty"`
}

type TriggerResponse struct {
	Storage map[string]*StorageLite `json:"storage,omitempty"`
}

func (r *RuntimesAPI) Trigger(*TriggerRequest) (*TriggerResponse, error) {
	return nil, status.ErrUnimplemented
}

type TriggerBatchRequest struct {
	Pipes []string `json:"pipes,omitempty"`
	Args  []*Args  `json:"args,omitempty"`
}

type TriggerBatchResponse struct {
	Storage []map[string]*StorageLite `json:"storage,omitempty"`
}

func (r *RuntimesAPI) TriggerBatch(*TriggerBatchRequest) (*TriggerBatchResponse, error) {
	return nil, status.ErrUnimplemented
}

type ImplicitTriggerRequest struct {
	URL   string `json:"url,omitempty"`
	Quiet bool   `json:"quiet,omitempty"`
}

type ImplicitTriggerResponse struct {
	Pipes   []*PipeLite             `json:"pipes,omitempty"`
	Storage map[string]*StorageLite `json:"storage,omitempty"`
}

func (r *RuntimesAPI) ImplicitTrigger(req *ImplicitTriggerRequest) (*ImplicitTriggerResponse, error) {
	r.api.mu.Lock()
	defer r.api.mu.Unlock()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.pipes.mu.Lock()
	defer r.pipes.mu.Unlock()

	triggered := []*PipeLite{}
	for _, p := range r.pipes.alwaysTrigger {
		triggered = append(triggered, p.PipeLite)
	}
	u, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", err, status.ErrInvalidArgument)
	}
	for _, p := range r.pipes.hostTrigger[u.Hostname()] {
		triggered = append(triggered, p.PipeLite)
	}
	// TODO: Do the actual trigger here.
	if req.Quiet {
		return &ImplicitTriggerResponse{}, nil
	}
	return &ImplicitTriggerResponse{Pipes: triggered}, nil
}

type RuntimeStatusRequest struct{}

type RuntimeStatusResponse struct{}

func (r *RuntimesAPI) RuntimeStatus(*RuntimeStatusRequest) (*RuntimeStatusResponse, error) {
	return &RuntimeStatusResponse{}, nil
}
