package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/wenooij/nuggit/status"
)

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

func ValidatePipe(p *Pipe) error {
	if p == nil {
		return fmt.Errorf("pipe is required: %w", status.ErrInvalidArgument)
	}
	if p.GetName() == "" {
		return fmt.Errorf("name is required: %w", status.ErrInvalidArgument)
	}
	for _, a := range p.Actions {
		// TODO: Validate actions.
		_ = a
	}
	// TODO: Validate point.
	return nil
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
	Pipe string `json:"pipe,omitempty"`
}

func (a *PipesAPI) validateCreatePipeRequest(req *CreatePipeRequest) error {
	if err := provided("pipe", "is", req.Pipe); err != nil {
		return err
	}
	if err := provided("actions", "are", req.Pipe.GetActions()); err != nil {
		return err
	}
	for i, action := range req.Pipe.Actions {
		if err := ValidateAction(&action, true /* = clientOnly */); err != nil {
			return fmt.Errorf("failed to validate action (#%d): %w", i, err)
		}
	}
	return nil
}

func (a *PipesAPI) CreatePipe(ctx context.Context, req *CreatePipeRequest) (*CreatePipeResponse, error) {
	if err := a.validateCreatePipeRequest(req); err != nil {
		return nil, err
	}
	pipeDigest, err := a.store.Store(ctx, req.Pipe)
	if err != nil {
		return nil, err
	}
	return &CreatePipeResponse{
		Pipe: pipeDigest,
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
	Pipes []string `json:"pipes,omitempty"`
}

func (a *PipesAPI) ListPipes(ctx context.Context, _ *ListPipesRequest) (*ListPipesResponse, error) {
	var res []string
	err := a.store.ScanRef(ctx, func(pipeDigest string, err error) error {
		if err != nil {
			return err
		}
		res = append(res, pipeDigest)
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
	if err := provided("pipe", "is", req.Pipe); err != nil {
		return nil, err
	}
	pipe, err := a.store.Load(ctx, req.Pipe)
	if err != nil {
		return nil, err
	}
	return &GetPipeResponse{Pipe: pipe}, nil
}

type GetPipesBatchRequest struct {
	IDs []string `json:"ids,omitempty"`
}

type GetPipesBatchResponse struct {
	Pipes   []*Pipe  `json:"pipes,omitempty"`
	Missing []string `json:"missing,omitempty"`
}

func (a *PipesAPI) GetPipesBatch(ctx context.Context, req *GetPipesBatchRequest) (*GetPipesBatchResponse, error) {
	if err := provided("ids", "are", req.IDs); err != nil {
		return nil, err
	}
	pipes, missing, err := a.store.LoadBatch(ctx, req.IDs)
	if err != nil {
		return nil, err
	}
	return &GetPipesBatchResponse{
		Pipes:   pipes,
		Missing: missing,
	}, nil
}
