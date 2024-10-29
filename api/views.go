package api

import (
	"context"
	"fmt"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/integrity"
	"github.com/wenooij/nuggit/status"
)

const viewsBaseURI = "/api/views"

func ValidateView(v *nuggit.View) error {
	if v == nil {
		return fmt.Errorf("view is required: %w", status.ErrInvalidArgument)
	}
	seen := make(map[integrity.NameDigest]struct{}, len(v.GetColumns()))
	for _, col := range v.GetColumns() {
		key, err := integrity.ParseNameDigest(col.Pipe)
		if err != nil {
			return err
		}
		if _, found := seen[key]; found {
			return fmt.Errorf("found duplicate column in view (%q; columns should be unique): %w", key, status.ErrInvalidArgument)
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
	View *nuggit.View `json:"view,omitempty"`
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
	if err := a.store.Store(ctx, ref.ID, *req.View); err != nil {
		return nil, err
	}
	return &CreateViewResponse{
		View: &ref,
	}, nil
}
