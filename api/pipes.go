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

func (p *PipeLite) GetRef() *Ref {
	if p == nil {
		return nil
	}
	return p.Ref
}

type PipeBase struct {
	Name    string    `json:"name,omitempty"`
	Actions []*Action `json:"actions,omitempty"`
	Point   *Point    `json:"point,omitempty"`
}

func (p *PipeBase) GetName() string {
	if p == nil {
		return ""
	}
	return p.Name
}

func (p *PipeBase) GetActions() []*Action {
	if p == nil {
		return nil
	}
	return p.Actions
}

type Pipe struct {
	*PipeLite `json:",omitempty"`
	*PipeBase `json:",omitempty"`
}

func (p *Pipe) GetLite() *PipeLite {
	if p == nil {
		return nil
	}
	return p.PipeLite
}

func (p *Pipe) GetBase() *PipeBase {
	if p == nil {
		return nil
	}
	return p.PipeBase
}

type PipesAPI struct {
	store PipeStorage
}

func (a *PipesAPI) Init(store PipeStorage) {
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
	Pipe *PipeBase `json:"pipe,omitempty"`
}

type CreatePipeResponse struct {
	Pipe *PipeLite `json:"pipe,omitempty"`
}

func (a *PipesAPI) CreatePipe(ctx context.Context, req *CreatePipeRequest) (*CreatePipeResponse, error) {
	if err := provided("pipe", "is", req.Pipe); err != nil {
		return nil, err
	}
	id, err := newUUID(ctx, a.store.Exists)
	if err != nil {
		return nil, err
	}
	pl := NewPipeLite(id)
	if err := a.store.Store(ctx, &Pipe{
		PipeLite: pl,
		PipeBase: req.Pipe,
	}); err != nil {
		return nil, err
	}
	return &CreatePipeResponse{
		Pipe: pl,
	}, nil
}

type ListPipesRequest struct{}

type ListPipesResponse struct {
	Pipes []*PipeLite `json:"pipes,omitempty"`
}

func (a *PipesAPI) ListPipes(ctx context.Context, _ *ListPipesRequest) (*ListPipesResponse, error) {
	n, _ := a.store.Len(ctx)
	res := make([]*PipeLite, 0, n)
	err := a.store.Scan(ctx, func(p *Pipe, err error) error {
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
