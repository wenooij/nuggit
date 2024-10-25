package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash"

	"github.com/wenooij/nuggit/status"
)

const pipesBaseURI = "/api/pipes"

type Pipe struct {
	NameDigest `json:"-"`
	Actions    []Action `json:"actions,omitempty"`
	Point      *Point   `json:"point,omitempty"`
}

func (p *Pipe) GetNameDigest() NameDigest {
	if p == nil {
		return NameDigest{}
	}
	return p.NameDigest
}

func (p *Pipe) GetName() string { return p.GetNameDigest().Name }

func (p *Pipe) SetNameDigest(nameDigest NameDigest) bool {
	if p == nil {
		return false
	}
	p.NameDigest = nameDigest
	return true
}

func (p *Pipe) GetActions() []Action {
	if p == nil {
		return nil
	}
	return p.Actions
}

func (p *Pipe) GetPoint() *Point {
	if p == nil {
		return nil
	}
	return p.Point
}

func (p *Pipe) writeDigest(h hash.Hash) error {
	return json.NewEncoder(h).Encode(p)
}

func ValidatePipe(p *Pipe, clientOnly bool) error {
	if p == nil {
		return fmt.Errorf("pipe is required: %w", status.ErrInvalidArgument)
	}
	if p.GetName() == "" {
		return fmt.Errorf("name is required: %w", status.ErrInvalidArgument)
	}
	for _, a := range p.Actions {
		if err := ValidateAction(a, clientOnly); err != nil {
			return err
		}
	}
	if err := ValidatePoint(p.Point); err != nil {
		return err
	}
	return nil
}

type PipesAPI struct {
	store    PipeStore
	criteria CriteriaStore
}

func (a *PipesAPI) Init(store PipeStore) {
	*a = PipesAPI{
		store: store,
	}
}

type DeletePipeRequest struct {
	Pipe *NameDigest `json:"pipe,omitempty"`
}

type DeletePipeResponse struct{}

func (a *PipesAPI) DeletePipe(ctx context.Context, req *DeletePipeRequest) (*DeletePipeResponse, error) {
	if err := provided("pipe", "is", req.Pipe); err != nil {
		return nil, err
	}
	if err := a.criteria.Disable(ctx, *req.Pipe); err != nil && !errors.Is(err, status.ErrNotFound) {
		return nil, err
	}
	return &DeletePipeResponse{}, nil
}

type CreatePipeRequest struct {
	*NameDigest `json:",omitempty"`
	Pipe        *Pipe `json:"pipe,omitempty"`
}

type CreatePipeResponse struct {
	Pipe *Ref `json:"pipe,omitempty"`
}

func (a *PipesAPI) CreatePipe(ctx context.Context, req *CreatePipeRequest) (*CreatePipeResponse, error) {
	if err := provided("name", "is", req.NameDigest); err != nil {
		return nil, err
	}
	if err := exclude("digest", "is", req.Digest); err != nil {
		return nil, err
	}
	if err := provided("pipe", "is", req.Pipe); err != nil {
		return nil, err
	}
	req.Pipe.NameDigest = *req.NameDigest
	if err := ValidatePipe(req.Pipe, true /* = clientOnly */); err != nil {
		return nil, err
	}
	if err := a.store.Store(ctx, req.Pipe); err != nil {
		return nil, err
	}

	var references []NameDigest
	for _, a := range req.Pipe.GetActions() {
		if a.GetAction() == "pipe" {
			references = append(references, a.GetNameDigestArg())
		}
	}

	ref := newNamedRef(pipesBaseURI, req.Pipe.NameDigest)
	return &CreatePipeResponse{
		Pipe: &ref,
	}, nil
}

type CreatePipesBatchRequest struct {
	Pipes []struct {
		NameDigest `json:",omitempty"`
		*Pipe      `json:",omitempty"`
	} `json:"pipes,omitempty"`
}

type CreatePipesBatchResponse struct {
	Pipes []Ref `json:"pipes,omitempty"`
}

func (a *PipesAPI) CreatePipesBatch(ctx context.Context, req *CreatePipesBatchRequest) (*CreatePipesBatchResponse, error) {
	if err := provided("pipes", "are", req.Pipes); err != nil {
		return nil, err
	}

	pipes := make([]*Pipe, 0, len(req.Pipes))
	for _, p := range req.Pipes {
		p.Pipe.SetNameDigest(p.NameDigest)
		pipes = append(pipes, p.Pipe)
	}

	if err := a.store.StoreBatch(ctx, pipes); err != nil {
		return nil, err
	}
	refs := make([]Ref, 0, len(req.Pipes))
	for _, pipe := range req.Pipes {
		refs = append(refs, newNamedRef(pipesBaseURI, pipe.NameDigest))
	}
	return &CreatePipesBatchResponse{Pipes: refs}, nil
}

type ListPipesRequest struct{}

type ListPipesResponse struct {
	Pipes []Ref `json:"pipes,omitempty"`
}

func (a *PipesAPI) ListPipes(ctx context.Context, _ *ListPipesRequest) (*ListPipesResponse, error) {
	var pipes []Ref
	for nameDigest, err := range a.store.ScanNames(ctx) {
		if err != nil {
			return nil, err
		}
		pipes = append(pipes, newNamedRef(pipesBaseURI, nameDigest))
	}
	return &ListPipesResponse{Pipes: pipes}, nil
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
	nameDigest, err := ParseNameDigest(req.Pipe)
	if err != nil {
		return nil, err
	}
	pipe, err := a.store.Load(ctx, nameDigest)
	if err != nil {
		return nil, err
	}
	return &GetPipeResponse{Pipe: pipe}, nil
}
