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
	Name        string `json:"name,omitempty"`
	Digest      string `json:"digest,omitempty"`
	nuggit.Pipe `json:",omitempty"`
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

func (p *Pipe) GetName() string {
	if p == nil {
		return ""
	}
	return p.Name
}

func (p *Pipe) GetDigest() string {
	if p == nil {
		return ""
	}
	return p.Digest
}

func (p *Pipe) SetName(name string) {
	if p != nil {
		p.Name = name
	}
}

func (p *Pipe) SetDigest(digest string) {
	if p != nil {
		p.Digest = digest
	}
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
	Name   string `json:"name,omitempty"`
	Digest string `json:"digest,omitempty"`
}

type DeletePipeResponse struct{}

func (a *PipesAPI) DeletePipe(ctx context.Context, req *DeletePipeRequest) (*DeletePipeResponse, error) {
	if err := provided("name", "is", req.Name); err != nil {
		return nil, err
	}
	if err := provided("digest", "is", req.Digest); err != nil {
		return nil, err
	}
	if err := a.rule.Disable(ctx, integrity.KeyLit(req.Name, req.Digest)); err != nil && !errors.Is(err, status.ErrNotFound) {
		return nil, err
	}
	return &DeletePipeResponse{}, nil
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
	if err := integrity.SetCheckDigest(req.Pipe, req.Pipe.Digest); err != nil {
		return nil, fmt.Errorf("failed to set digest (%q): %w", req.Pipe.GetName(), err)
	}
	if err := ValidatePipe(req.Pipe, true /* = clientOnly */); err != nil {
		return nil, err
	}
	if err := a.store.Store(ctx, req.Pipe); err != nil {
		return nil, err
	}

	ref := newNamedRef(pipesBaseURI, req.Pipe)
	return &CreatePipeResponse{Pipe: &ref}, nil
}

type CreatePipesBatchRequest struct {
	Pipes []*Pipe `json:"pipes,omitempty"`
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
		if err := integrity.SetCheckDigest(pipe, p.Digest); err != nil {
			return nil, fmt.Errorf("failed to set digest (%q): %w", p.GetName(), err)
		}
		pipes = append(pipes, pipe)
	}

	if err := a.store.StoreBatch(ctx, pipes); err != nil {
		return nil, err
	}
	refs := make([]Ref, 0, len(req.Pipes))
	for _, pipe := range req.Pipes {
		refs = append(refs, newNamedRef(pipesBaseURI, pipe))
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
