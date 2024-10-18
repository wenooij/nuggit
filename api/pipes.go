package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/wenooij/nuggit/status"
)

func NewPipeRef(id string) *Ref {
	return newRef("/api/pipes/", id)
}

type Pipe struct {
	Name    string   `json:"name,omitempty"`
	Actions []Action `json:"actions,omitempty"`
	Point   *Point   `json:"point,omitempty"`
}

func (p *Pipe) GetName() string {
	if p == nil {
		return ""
	}
	return p.Name
}

func (p *Pipe) GetActions() []Action {
	if p == nil {
		return nil
	}
	return p.Actions
}

type PipesAPI struct {
	store PipeStore
}

func (a *PipesAPI) Init(store PipeStore) {
	*a = PipesAPI{
		store: store,
	}
}

type DeletePipeRequest struct {
	Pipe string `json:"pipe,omitempty"`
}

type DeletePipeResponse struct{}

func (a *PipesAPI) DeletePipe(ctx context.Context, req *DeletePipeRequest) (*DeletePipeResponse, error) {
	if err := a.store.Delete(ctx, req.Pipe); err != nil && !errors.Is(err, status.ErrNotFound) {
		return nil, err
	}
	return &DeletePipeResponse{}, nil
}

type DeletePipeRequestBatch struct {
	Pipes []string `json:"pipes,omitempty"`
}

type DeletePipeResponseBatch struct{}

func (r *PipesAPI) DeleteBatch(*DeletePipeRequestBatch) (*DeletePipeResponseBatch, error) {
	return nil, fmt.Errorf("not implemented")
}

type CreatePipeRequest struct {
	Pipe *Pipe `json:"pipe,omitempty"`
}

type CreatePipeResponse struct {
	Pipe *Ref `json:"pipe,omitempty"`
}

func (a *PipesAPI) CreatePipe(ctx context.Context, req *CreatePipeRequest) (*CreatePipeResponse, error) {
	if err := provided("pipe", "is", req.Pipe); err != nil {
		return nil, err
	}
	id, err := a.store.Store(ctx, req.Pipe)
	if err != nil {
		return nil, err
	}
	return &CreatePipeResponse{
		Pipe: NewPipeRef(id),
	}, nil
}

type CreatePipesBatchRequest struct {
	Pipes []*Pipe `json:"pipes,omitempty"`
}

type CreatePipesBatchResponse struct {
	Pipes []*Ref `json:"pipes,omitempty"`
}

func (a *PipesAPI) CreatePipesBatch(ctx context.Context, req *CreatePipesBatchRequest) (*CreatePipesBatchResponse, error) {
	return nil, status.ErrUnimplemented
}

type ListPipesRequest struct{}

type ListPipesResponse struct {
	Pipes []*Ref `json:"pipes,omitempty"`
}

func (a *PipesAPI) ListPipes(ctx context.Context, _ *ListPipesRequest) (*ListPipesResponse, error) {
	var res []*Ref
	err := a.store.ScanRef(ctx, func(id string, err error) error {
		if err != nil {
			return err
		}
		res = append(res, NewPipeRef(id))
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
	Pipe *Pipe `json:"pipe,omitempty"`
}

func (a *PipesAPI) GetPipe(ctx context.Context, req *GetPipeRequest) (*GetPipeResponse, error) {
	pipe, err := a.store.Load(ctx, req.Pipe)
	if err != nil {
		return nil, err
	}
	return &GetPipeResponse{Pipe: pipe}, nil
}

type GetPipesBatchRequest struct {
	Pipes []string `json:"pipes,omitempty"`
}

type GetPipesBatchResponse struct {
	Pipes   []*Pipe  `json:"pipes,omitempty"`
	Missing []string `json:"missing,omitempty"`
}

func (a *PipesAPI) GetPipesBatch(ctx context.Context, req *GetPipesBatchRequest) (*GetPipesBatchResponse, error) {
	pipes := make([]*Pipe, 0, len(req.Pipes))
	var missing []string
	for _, id := range req.Pipes {
		pipe, err := a.store.Load(ctx, id)
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
