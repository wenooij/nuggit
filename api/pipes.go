package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/integrity"
	"github.com/wenooij/nuggit/status"
)

const pipesBaseURI = "/api/pipes"

type Pipe struct {
	integrity.NameDigest `json:",omitempty"`
	nuggit.Pipe          `json:",omitempty"`
}

func (p *Pipe) GetActions() []nuggit.Action {
	if p == nil {
		return nil
	}
	return p.Actions
}

func (p *Pipe) GetPoint() nuggit.Point {
	if p == nil {
		return nuggit.Point{}
	}
	return p.Point
}

func (p *Pipe) GetNameDigest() integrity.NameDigest {
	if p == nil {
		return integrity.NameDigest{}
	}
	return p.NameDigest
}

func (p *Pipe) GetName() string { return p.GetNameDigest().Name }

func (p *Pipe) GetDigest() string { return p.GetNameDigest().Digest }

func (p *Pipe) SetNameDigest(nameDigest integrity.NameDigest) bool {
	if p == nil {
		return false
	}
	p.NameDigest = nameDigest
	return true
}

var supportedScalars = map[nuggit.Scalar]struct{}{
	"":             {}, // Same as Bytes.
	nuggit.Bytes:   {},
	nuggit.String:  {},
	nuggit.Bool:    {},
	nuggit.Int64:   {},
	nuggit.Uint64:  {},
	nuggit.Float64: {},
}

func ValidateScalar(s nuggit.Scalar) error {
	_, ok := supportedScalars[s]
	if !ok {
		return fmt.Errorf("scalar type is not supported (%q)", s)
	}
	return nil
}

func ValidatePoint(p nuggit.Point) error {
	// Scalar == "" is valid and equivalent to bytes.
	return ValidateScalar(p.Scalar)
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
	store PipeStore
	rule  RuleStore
}

func (a *PipesAPI) Init(store PipeStore, rule RuleStore) {
	*a = PipesAPI{
		store: store,
		rule:  rule,
	}
}

type DeletePipeRequest struct {
	Pipe *integrity.NameDigest `json:"pipe,omitempty"`
}

type DeletePipeResponse struct{}

func (a *PipesAPI) DeletePipe(ctx context.Context, req *DeletePipeRequest) (*DeletePipeResponse, error) {
	if err := provided("pipe", "is", req.Pipe); err != nil {
		return nil, err
	}
	if err := a.rule.Disable(ctx, *req.Pipe); err != nil && !errors.Is(err, status.ErrNotFound) {
		return nil, err
	}
	return &DeletePipeResponse{}, nil
}

type CreatePipeRequest struct {
	*integrity.NameDigest `json:",omitempty"`
	Pipe                  *Pipe `json:"pipe,omitempty"`
}

type CreatePipeResponse struct {
	Pipe *Ref `json:"pipe,omitempty"`
}

func (a *PipesAPI) CreatePipe(ctx context.Context, req *CreatePipeRequest) (*CreatePipeResponse, error) {
	if err := provided("name", "is", req.NameDigest); err != nil {
		return nil, err
	}
	if err := exclude("digest", "is", req.Digest); err != nil {
		// TODO: Instead of excluding digest, verify the digest here.
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

	nameDigest, err := integrity.NewNameDigest(req.Pipe)
	if err != nil {
		return nil, err
	}

	ref := newNamedRef(pipesBaseURI, nameDigest)
	return &CreatePipeResponse{Pipe: &ref}, nil
}

type CreatePipesBatchRequest struct {
	Pipes []struct {
		integrity.NameDigest `json:",omitempty"`
		nuggit.Pipe          `json:",omitempty"`
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
		pipe := new(Pipe)
		pipe.Pipe = p.Pipe
		pipe.SetNameDigest(p.NameDigest)
		pipes = append(pipes, pipe)
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
	nameDigest, err := integrity.ParseNameDigest(req.Pipe)
	if err != nil {
		return nil, err
	}
	pipe, err := a.store.Load(ctx, nameDigest)
	if err != nil {
		return nil, err
	}
	return &GetPipeResponse{Pipe: pipe}, nil
}
