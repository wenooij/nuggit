package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"iter"
	"maps"
	"slices"

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

func (p *Pipe) SetNameDigest(name NameDigest) {
	if p != nil {
		p.NameDigest = name
	}
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

// FlattenPipe recursively replaces all pipe actions with their definitions
// returning a new Pipe or an error if the process failed.
// The flattened pipe is fully hermetric, making no references to other pipes.
//
// NOTE: The returned pipe will have a different digest than the input pipe.
//
// TODO: check the digests of pipes in referencedPipes.
func FlattenPipe(referencedPipes map[NameDigest]*Pipe, pipe *Pipe) (*Pipe, error) {
	actions := slices.Clone(pipe.GetActions())
	for i := 0; i < len(actions); {
		a := actions[i]
		if a.GetAction() != "pipe" {
			i++
			continue
		}
		pipe := a.GetNameDigestArg()
		referencedPipe, ok := referencedPipes[pipe]
		if !ok {
			return nil, fmt.Errorf("referenced pipe not found (%q): %w", &pipe, status.ErrInvalidArgument)
		}
		actions = slices.Insert(slices.Delete(actions, i, i+1), i, referencedPipe.GetActions()...)
	}
	pipe = &Pipe{
		Actions: actions,
		Point:   pipe.GetPoint(),
		NameDigest: NameDigest{
			Name: pipe.GetName(),
		},
	}
	nameDigest, err := NewNameDigest(pipe)
	if err != nil {
		return nil, err
	}
	pipe.SetNameDigest(nameDigest)
	return pipe, nil
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
	Pipe *NameDigest `json:"pipe,omitempty"`
}

type DeletePipeResponse struct{}

func (a *PipesAPI) DeletePipe(ctx context.Context, req *DeletePipeRequest) (*DeletePipeResponse, error) {
	if err := provided("pipe", "is", req.Pipe); err != nil {
		return nil, err
	}
	if err := a.store.Delete(ctx, *req.Pipe); err != nil && !errors.Is(err, status.ErrNotFound) {
		return nil, err
	}
	return &DeletePipeResponse{}, nil
}

type DeletePipeRequestBatch struct {
	Pipes []NameDigest `json:"pipes,omitempty"`
}

type DeletePipeResponseBatch struct{}

func (r *PipesAPI) DeleteBatch(*DeletePipeRequestBatch) (*DeletePipeResponseBatch, error) {
	return nil, fmt.Errorf("not implemented")
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
	nameDigest, err := a.store.Store(ctx, req.Pipe)
	if err != nil {
		return nil, err
	}

	var references []NameDigest
	for _, a := range req.Pipe.GetActions() {
		if a.GetAction() == "pipe" {
			references = append(references, a.GetNameDigestArg())
		}
	}
	if err := a.store.StorePipeReferences(ctx, nameDigest, references); err != nil {
		return nil, err
	}

	ref := newNamedRef(pipesBaseURI, nameDigest)
	return &CreatePipeResponse{
		Pipe: &ref,
	}, nil
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
	pipes, err := a.store.StoreBatch(ctx, req.Pipes)
	if err != nil {
		return nil, err
	}
	var refs []Ref
	for _, name := range pipes {
		refs = append(refs, newNamedRef(pipesBaseURI, name))
	}
	return &CreatePipesBatchResponse{Pipes: refs}, nil
}

type ListPipesRequest struct{}

type ListPipesResponse struct {
	Pipes []Ref `json:"pipes,omitempty"`
}

func (a *PipesAPI) ListPipes(ctx context.Context, _ *ListPipesRequest) (*ListPipesResponse, error) {
	var pipes []Ref
	for name, err := range a.store.ScanNames(ctx) {
		if err != nil {
			return nil, err
		}
		pipes = append(pipes, newNamedRef(pipesBaseURI, name))
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

type GetPipesBatchRequest struct {
	Pipes []NameDigest `json:"pipes,omitempty"`
}

type GetPipesBatchResponse struct {
	Pipes   []*Pipe      `json:"pipes,omitempty"`
	Missing []NameDigest `json:"missing,omitempty"`
}

func (a *PipesAPI) GetPipesBatch(ctx context.Context, req *GetPipesBatchRequest) (*GetPipesBatchResponse, error) {
	if err := provided("pipes", "are", req.Pipes); err != nil {
		return nil, err
	}
	next, stop := iter.Pull(slices.Values(req.Pipes))
	remaining := maps.Collect(func(yield func(k NameDigest, v struct{}) bool) {
		for k, ok := next(); ok && yield(k, struct{}{}); k, ok = next() {
		}
		stop()
	})
	var pipes []*Pipe
	for pipe, err := range a.store.LoadBatch(ctx, req.Pipes) {
		if err != nil {
			return nil, err
		}
		pipes = append(pipes, pipe)
		delete(remaining, pipe.GetNameDigest())
	}
	return &GetPipesBatchResponse{
		Pipes:   pipes,
		Missing: slices.Collect(maps.Keys(remaining)),
	}, nil
}
