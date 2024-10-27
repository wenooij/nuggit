package api

import (
	"context"
	"encoding/json"
	"fmt"
	"hash"

	"github.com/wenooij/nuggit/integrity"
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

func ValidateView(c *View) error {
	if c == nil {
		return fmt.Errorf("view is required: %w", status.ErrInvalidArgument)
	}
	seen := make(map[integrity.NameDigest]struct{}, len(c.GetColumns()))
	for _, col := range c.GetColumns() {
		key := integrity.Key(col.Pipe)
		if _, found := seen[key]; found {
			return fmt.Errorf("found duplicate pipe@digest in view (%q; pipes should be unique): %w", key, status.ErrInvalidArgument)
		}
		seen[key] = struct{}{}
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
