package api

import (
	"context"
	"encoding/json"
	"fmt"
	"hash"

	"github.com/wenooij/nuggit/status"
)

const viewsBaseURI = "/api/views"

type View struct {
	Alias   string       `json:"alias,omitempty"`
	Columns []ViewColumn `json:"columns,omitempty"`
}

func (c *View) GetAlias() string {
	if c == nil {
		return ""
	}
	return c.Alias
}

func (c *View) GetColumns() []ViewColumn {
	if c == nil {
		return nil
	}
	return c.Columns
}

func (c *View) WriteDigest(h hash.Hash) error {
	return json.NewEncoder(h).Encode(c)
}

type ViewColumn struct {
	Alias string `json:"alias,omitempty"`
	Pipe  *Pipe  `json:"pipe,omitempty"`
}

func (c ViewColumn) GetNameDigest() NameDigest {
	return c.Pipe.GetNameDigest()
}

func ValidateView(c *View) error {
	if c == nil {
		return fmt.Errorf("view is required: %w", status.ErrInvalidArgument)
	}
	seen := make(map[NameDigest]struct{}, len(c.GetColumns()))
	for _, col := range c.GetColumns() {
		if _, found := seen[col.GetNameDigest()]; found {
			return fmt.Errorf("found duplicate pipe@digest in view (%q; pipes should be unique): %w", col, status.ErrInvalidArgument)
		}
		seen[col.GetNameDigest()] = struct{}{}
	}
	return nil
}

func ValidateViewPipes(c *View, pipes []*Pipe) error {
	if err := ValidateView(c); err != nil {
		return err
	}
	expected := make(map[NameDigest]struct{}, len(c.GetColumns()))
	for _, col := range c.GetColumns() {
		expected[col.GetNameDigest()] = struct{}{}
	}
	seen := make(map[NameDigest]struct{}, len(pipes))
	for _, p := range pipes {
		if err := ValidatePipe(p, true /* = clientOnly */); err != nil {
			return err
		}
		nameDigest, err := NewNameDigest(p)
		if err != nil {
			return err
		}
		if _, found := seen[nameDigest]; found {
			return fmt.Errorf("found duplicate name@digest in request context (%q; pipes should be unique): %w", nameDigest, status.ErrInvalidArgument)
		}
		seen[nameDigest] = struct{}{}
		if _, found := expected[nameDigest]; !found {
			return fmt.Errorf("mismatch in name@digest from view and request context (%q): %w", nameDigest, status.ErrInvalidArgument)
		}
		delete(expected, nameDigest)
	}
	if err := CheckIntegrity(c.GetColumns(), pipes); err != nil {
		return err
	}
	return nil
}

func ValidateViewPipesSubset(c *View, pipes []*Pipe) error {
	if err := ValidateView(c); err != nil {
		return err
	}
	allowed := make(map[NameDigest]struct{}, len(c.GetColumns()))
	for _, col := range c.GetColumns() {
		allowed[col.GetNameDigest()] = struct{}{}
	}
	if err := CheckIntegritySubset(allowed, pipes); err != nil {
		return err
	}
	return nil
}

type ViewsAPI struct {
	store ViewStore
	pipes PipeStore
}

func (a *ViewsAPI) Init(store ViewStore, pipes PipeStore) {
	*a = ViewsAPI{
		store: store,
		pipes: pipes,
	}
}

type CreateViewRequest struct {
	View *View `json:"view,omitempty"`
}

type CreateViewResponse struct {
	View *Ref `json:"view,omitempty"`
}

func (a *ViewsAPI) CreateView(ctx context.Context, req *CreateViewRequest) (*CreateViewResponse, error) {
	if err := provided("view", "is", req.View); err != nil {
		return nil, err
	}
	if err := ValidateView(req.View); err != nil {
		return nil, err
	}
	ref, err := newRef(viewsBaseURI)
	if err != nil {
		return nil, err
	}
	if err := a.store.Store(ctx, ref.ID, req.View); err != nil {
		return nil, err
	}
	return &CreateViewResponse{
		View: &ref,
	}, nil
}
