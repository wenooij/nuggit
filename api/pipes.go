package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/wenooij/nuggit/status"
)

type PipeLite struct {
	*Ref `json:",omitempty"`
}

func NewPipeLite(id string) *PipeLite {
	return &PipeLite{newRef("/api/pipes/", id)}
}

type PipeBase struct {
	Sequence []*NodeLite `json:"sequence,omitempty"`
}

type Pipe struct {
	*PipeLite `json:",omitempty"`
	*PipeBase `json:",omitempty"`
}

type PipeState struct {
	Collections map[string]struct{} `json:"collections,omitempty"`
}

type PipeBaseRich struct {
	Sequence []*NodeBase `json:"sequence,omitempty"`
}

type PipeRich struct {
	*PipeLite `json:",omitempty"`
	Sequence  []*NodeLite `json:"sequence,omitempty"`
	State     *PipeState  `json:"state,omitempty"`
}

type PipesAPI struct {
	nodes   *NodesAPI
	storage PipeStorage
}

func (a *PipesAPI) Init(store PipeStorage, nodes *NodesAPI) {
	*a = PipesAPI{
		nodes:   nodes,
		storage: store,
	}
}

type DeletePipeRequest struct {
	ID        string `json:"id,omitempty"`
	KeepNodes bool   `json:"keep_nodes,omitempty"`
}

type DeletePipeResponse struct{}

func (a *PipesAPI) DeletePipe(ctx context.Context, req *DeletePipeRequest) (*DeletePipeResponse, error) {
	if !req.KeepNodes {
		pipe, err := a.storage.Load(ctx, req.ID)
		if err != nil {
			if errors.Is(err, status.ErrNotFound) {
				return &DeletePipeResponse{}, nil
			}
			return nil, err
		}
		for _, n := range pipe.Sequence {
			if _, err := a.nodes.DeleteNode(ctx, &DeleteNodeRequest{ID: n.UUID()}); err != nil && !errors.Is(err, status.ErrNotFound) {
				return nil, err
			}
		}
	}
	if err := a.storage.Delete(ctx, req.ID); err != nil {
		if !errors.Is(err, status.ErrNotFound) {
			return nil, err
		}
	}
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
	Pipe  *PipeLite   `json:"pipe,omitempty"`
	Nodes []*NodeLite `json:"nodes,omitempty"`
}

func (a *PipesAPI) CreatePipe(ctx context.Context, req *CreatePipeRequest) (*CreatePipeResponse, error) {
	if err := provided("pipe", "is", req.Pipe); err != nil {
		return nil, err
	}
	id, err := newUUID(func(id string) error { _, err := a.storage.Load(ctx, id); return err })
	if err != nil {
		return nil, err
	}
	pl := NewPipeLite(id)
	seq := make([]*NodeLite, 0, len(req.Pipe.Sequence))
	for _, n := range req.Pipe.Sequence {
		id, err := newUUID(func(id string) error { _, err := a.nodes.loadNode(ctx, id); return err })
		if err != nil {
			return nil, err
		}
		nl := NewNodeLite(id)
		node := &NodeRich{
			Node: &Node{
				NodeLite: nl,
				NodeBase: n,
			},
		}
		a.nodes.createNode(ctx, node) // createNode always returns true.
		seq = append(seq, nl)
	}
	if err := a.storage.Store(ctx, &PipeRich{
		PipeLite: pl,
		Sequence: seq,
	}); err != nil {
		return nil, err
	}
	return &CreatePipeResponse{
		Pipe:  pl,
		Nodes: seq,
	}, nil
}

type ListPipesRequest struct{}

type ListPipesResponse struct {
	Pipes []*PipeLite `json:"pipes,omitempty"`
}

func (a *PipesAPI) ListPipes(ctx context.Context, _ *ListPipesRequest) (*ListPipesResponse, error) {
	n, _ := a.storage.Len(ctx)
	res := make([]*PipeLite, 0, n)
	err := a.storage.Scan(ctx, func(p *PipeRich, err error) error {
		if err != nil {
			return err
		}
		res = append(res, p.PipeLite)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &ListPipesResponse{Pipes: res}, nil
}

type GetPipeRequest struct {
	Pipe string `json:"pipe,omitempty"`
}

type GetPipeResponse struct {
	Pipe *PipeRich `json:"pipe,omitempty"`
}

func (a *PipesAPI) GetPipe(ctx context.Context, req *GetPipeRequest) (*GetPipeResponse, error) {
	pipe, err := a.storage.Load(ctx, req.Pipe)
	if err != nil {
		return nil, err
	}
	return &GetPipeResponse{Pipe: pipe}, nil
}

type GetPipesBatchRequest struct {
	Pipes []string `json:"pipes,omitempty"`
}

type GetPipesBatchResponse struct {
	Pipes   []*PipeRich `json:"pipes,omitempty"`
	Missing []string    `json:"missing,omitempty"`
}

func (a *PipesAPI) GetPipesBatch(ctx context.Context, req *GetPipesBatchRequest) (*GetPipesBatchResponse, error) {
	pipes := make([]*PipeRich, 0, len(req.Pipes))
	var missing []string
	for _, id := range req.Pipes {
		pipe, err := a.storage.Load(ctx, id)
		if err != nil {
			if !errors.Is(err, status.ErrNotFound) {
				return nil, err
			}
			missing = append(missing, id)
		}
		pipes = append(pipes, pipe)
	}
	return &GetPipesBatchResponse{
		Pipes:   pipes,
		Missing: missing,
	}, nil
}
