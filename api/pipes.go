package api

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/wenooij/nuggit/status"
)

type PipeLite struct {
	*Ref `json:",omitempty"`
}

type PipeBase struct {
	Sequence   []*NodeLite `json:"sequence,omitempty"`
	Conditions *Conditions `json:"conditions,omitempty"`
	State      *PipeState  `json:"state,omitempty"`
}

type Pipe struct {
	*PipeLite `json:",omitempty"`
	*PipeBase `json:",omitempty"`
}

type PipeState struct {
	Disabled bool `json:"disabled,omitempty"`
}

type PipeBaseRich struct {
	Sequence   []*NodeBase `json:"sequence,omitempty"`
	Conditions *Conditions `json:"conditions,omitempty"`
	State      *PipeState  `json:"state,omitempty"`
}

type PipeRich struct {
	*PipeLite     `json:",omitempty"`
	*PipeBaseRich `json:",omitempty"`
}

type Conditions struct {
	AlwaysTriggered bool   `json:"always_triggered,omitempty"`
	Host            string `json:"host,omitempty"`
	URLPattern      string `json:"url_pattern,omitempty"`
}

type Args struct {
	Elements []DOMElement `json:"elements,omitempty"`
	Bytes    [][]byte     `json:"bytes,omitempty"`
}

type PipesAPI struct {
	api            *API
	nodes          *NodesAPI
	pipes          map[string]*Pipe            // pipe ID => Pipe.
	alwaysTrigger  map[string]*Pipe            // pipe ID => {}.
	hostTrigger    map[string]map[string]*Pipe // host name => pipe ID => Pipe.
	patternTrigger map[string]*regexp.Regexp   // pipe ID => URL pattern.
	mu             sync.RWMutex
}

func (a *PipesAPI) Init(api *API, nodes *NodesAPI) {
	*a = PipesAPI{
		api:            api,
		nodes:          nodes,
		pipes:          make(map[string]*Pipe),
		alwaysTrigger:  make(map[string]*Pipe),
		hostTrigger:    make(map[string]map[string]*Pipe),
		patternTrigger: make(map[string]*regexp.Regexp),
	}
}

// locks excluded: api.mu, mu, nodes.mu.
func (a *PipesAPI) deletePipe(pipeID string, keepNodes bool) {
	pipe, ok := a.pipes[pipeID]
	if !ok {
		return
	}
	for _, node := range pipe.Sequence {
		a.nodes.deletePipeNode(pipeID, node.ID, keepNodes)
	}
	delete(a.pipes, pipeID)
}

type DeletePipeRequest struct {
	ID        string `json:"id,omitempty"`
	KeepNodes bool   `json:"keep_nodes,omitempty"`
}

type DeletePipeResponse struct{}

func (r *PipesAPI) DeletePipe(req *DeletePipeRequest) (*DeletePipeResponse, error) {
	r.api.mu.Lock()
	defer r.api.mu.Unlock()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodes.mu.Lock()
	defer r.nodes.mu.Unlock()

	r.deletePipe(req.ID, req.KeepNodes)
	return &DeletePipeResponse{}, nil
}

type DeletePipeRequestBatch struct {
	Names []string
}

type DeletePipeResponseBatch struct{}

func (r *PipesAPI) DeleteBatch(*DeletePipeRequestBatch) (*DeletePipeResponseBatch, error) {
	return nil, fmt.Errorf("not implemented")
}

type CreatePipeRequest struct {
	Pipe *PipeBaseRich `json:"pipe,omitempty"`
}

type CreatePipeResponse struct {
	Pipe *PipeLite `json:"pipe,omitempty"`
}

func (r *PipesAPI) CreatePipe(req *CreatePipeRequest) (*CreatePipeResponse, error) {
	r.api.mu.Lock()
	defer r.api.mu.Unlock()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodes.mu.Lock()
	defer r.nodes.mu.Unlock()

	id, err := newUUID(func(id string) bool { return r.pipes[id] == nil })
	if err != nil {
		return nil, err
	}
	pl := &PipeLite{
		Ref: &Ref{
			ID:  id,
			URI: fmt.Sprintf("/api/pipes/%s", id),
		},
	}

	seq := make([]*NodeLite, 0, len(req.Pipe.Sequence))
	for _, n := range req.Pipe.Sequence {
		id, err := newUUID(func(id string) bool { return r.nodes.nodes[id] == nil })
		if err != nil {
			return nil, err
		}
		nl := &NodeLite{
			Ref: &Ref{
				ID:  id,
				URI: fmt.Sprintf("/api/nodes/%s", id),
			},
		}
		node := &Node{
			NodeLite: nl,
			NodeBase: n,
		}
		r.nodes.createNode(node) // createNode always returns true.
		seq = append(seq, nl)
	}
	pipe := &Pipe{
		PipeLite: pl,
		PipeBase: &PipeBase{
			Conditions: req.Pipe.Conditions,
			Sequence:   seq,
		},
	}
	r.pipes[id] = pipe
	return &CreatePipeResponse{Pipe: pl}, nil
}

type ListPipesRequest struct{}

type ListPipesResponse struct {
	Pipes []*PipeLite `json:"pipes,omitempty"`
}

func (r *PipesAPI) ListPipes(*ListPipesRequest) (*ListPipesResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	res := make([]*PipeLite, 0, len(r.pipes))
	for _, p := range r.pipes {
		res = append(res, p.PipeLite)
	}
	return &ListPipesResponse{Pipes: res}, nil
}

type GetPipeRequest struct {
	Pipe string `json:"pipe,omitempty"`
}

type GetPipeResponse struct {
	Pipe *Pipe `json:"pipe,omitempty"`
}

func (r *PipesAPI) GetPipe(req *GetPipeRequest) (*GetPipeResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	pipe, ok := r.pipes[req.Pipe]
	if !ok {
		return nil, fmt.Errorf("failed to load pipe: %w", status.ErrNotFound)
	}
	return &GetPipeResponse{Pipe: pipe}, nil
}
