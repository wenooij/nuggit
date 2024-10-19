package api

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

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

func PipeDigestSHA1(p *Pipe) (string, error) {
	h := sha1.New()
	data, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	if _, err := h.Write(data); err != nil {
		return "", err
	}
	digest := h.Sum(nil)
	return hex.EncodeToString(digest), nil
}

var namePattern = regexp.MustCompile(`^(?i:[a-z][a-z0-9-]*)$`)

func validatePipeName(name string) error {
	if name == "" {
		return fmt.Errorf("name must not be empty: %w", status.ErrInvalidArgument)
	}
	if !namePattern.MatchString(name) {
		return fmt.Errorf("name contains invalid characters (%q): %w", name, status.ErrInvalidArgument)
	}
	return nil
}

func validateHexDigest(hexStr string) error {
	for _, b := range hexStr {
		switch {
		case b >= '0' && b <= '9' || b >= 'A' && b <= 'F' || b >= 'a' && b <= 'f':
		default:
			return fmt.Errorf("digest is not hex encoded (%q): %v", hexStr, status.ErrInvalidArgument)
		}
	}
	return nil
}

func JoinPipeDigest(name string, digest string) (string, error) {
	if len(digest) == 0 {
		return "", fmt.Errorf("digest must not be empty: %w", status.ErrInvalidArgument)
	}
	if err := validatePipeName(name); err != nil {
		return "", err
	}
	if err := validateHexDigest(digest); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s@%s", name, digest), nil
}

func SplitPipeDigest(pipeDigest string) (string, string, error) {
	if len(pipeDigest) == 0 {
		return "", "", fmt.Errorf("pipe@digest must not be empty: %w", status.ErrInvalidArgument)
	}
	elems := strings.Split(pipeDigest, "@")
	if len(elems) != 2 {
		return "", "", fmt.Errorf("pipe@digest is missing '@' delimiter (%q): %w", pipeDigest, status.ErrInvalidArgument)
	}
	name := elems[0]
	if err := validatePipeName(name); err != nil {
		return "", "", err
	}
	digest := elems[1]
	if err := validateHexDigest(digest); err != nil {
		return "", "", err
	}
	return name, digest, nil
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

func (a *PipesAPI) validateAction(action Action, allowPipe bool) error {
	if err := validateAction(&action); err != nil {
		return err
	}
	if action.GetAction() == ActionExchange {
		return fmt.Errorf("exchange is not allowed here: %w", status.ErrInvalidArgument)
	}
	if !allowPipe && action.GetAction() == ActionPipe {
		return fmt.Errorf("pipe action is not supported here as its references cannot be hermetically verified (try /api/pipes/batch instead): %w", status.ErrInvalidArgument)
	}
	return nil
}

func (a *PipesAPI) validateCreatePipeRequest(req *CreatePipeRequest) error {
	if err := provided("pipe", "is", req.Pipe); err != nil {
		return err
	}
	if err := provided("actions", "are", req.Pipe.GetActions()); err != nil {
		return err
	}
	for i, action := range req.Pipe.Actions {
		if err := a.validateAction(action, false /* = allowPipe */); err != nil {
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
