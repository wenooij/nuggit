package api

import (
	"context"
)

func NewCollectionRef(id string) *Ref {
	return newRef("/api/collections/", id)
}

type CollectionConditions struct {
	AlwaysTrigger bool   `json:"always_trigger,omitempty"`
	Hostname      string `json:"hostname,omitempty"`
	URLPattern    string `json:"url_pattern,omitempty"`
}

func (c *CollectionConditions) GetAlwaysTrigger() bool {
	if c == nil {
		return false
	}
	return c.AlwaysTrigger
}

func (c *CollectionConditions) GetHostname() string {
	if c == nil {
		return ""
	}
	return c.Hostname
}

func (c *CollectionConditions) GetURLPattern() string {
	if c == nil {
		return ""
	}
	return c.URLPattern
}

type Collection struct {
	Name       string                `json:"name,omitempty"`
	Pipes      []string              `json:"pipes,omitempty"`
	Conditions *CollectionConditions `json:"conditions,omitempty"`
}

func (c *Collection) GetName() string {
	if c == nil {
		return ""
	}
	return c.Name
}

func (c *Collection) GetPipes() []string {
	if c == nil {
		return nil
	}
	return c.Pipes
}

func (c *Collection) GetConditions() *CollectionConditions {
	if c == nil {
		return nil
	}
	return c.Conditions
}

type CollectionData struct {
	Collection string `json:"collection,omitempty"`
	Values     []any  `json:"values,omitempty"`
}

type CollectionsAPI struct {
	store CollectionStore
}

func (a *CollectionsAPI) Init(store CollectionStore) {
	*a = CollectionsAPI{
		store: store,
	}
}

type CreateCollectionRequest struct {
	Collection *Collection `json:"collection,omitempty"`
}

type CreateCollectionResponse struct {
	Collection *Ref `json:"collection,omitempty"`
}

func (a *CollectionsAPI) CreateCollection(ctx context.Context, req *CreateCollectionRequest) (*CreateCollectionResponse, error) {
	if err := provided("collection", "is", req.Collection); err != nil {
		return nil, err
	}
	if err := provided("name", "is", req.Collection.Name); err != nil {
		return nil, err
	}
	id, err := a.store.Store(ctx, req.Collection)
	if err != nil {
		return nil, err
	}
	return &CreateCollectionResponse{
		Collection: NewCollectionRef(id),
	}, nil
}

type GetCollectionRequest struct {
	Collection string `json:"collection,omitempty"`
}

type GetCollectionResponse struct {
	Collection *Collection `json:"collection,omitempty"`
}

func (a *CollectionsAPI) GetCollection(ctx context.Context, req *GetCollectionRequest) (*GetCollectionResponse, error) {
	if err := provided("collection", "is", req.Collection); err != nil {
		return nil, err
	}
	collection, err := a.store.Load(ctx, req.Collection)
	if err != nil {
		return nil, err
	}
	return &GetCollectionResponse{Collection: collection}, nil
}

type ListCollectionsRequest struct{}

type ListCollectionsResponse struct {
	Collections []*Ref `json:"collections,omitempty"`
}

func (a *CollectionsAPI) ListCollections(ctx context.Context, req *ListCollectionsRequest) (*ListCollectionsResponse, error) {
	var res []*Ref
	if err := a.store.ScanRef(ctx, func(id string, err error) error {
		if err != nil {
			return err
		}
		res = append(res, NewCollectionRef(id))
		return nil
	}); err != nil {
		return nil, err
	}
	return &ListCollectionsResponse{Collections: res}, nil
}

type DeleteCollectionRequest struct {
	Collection string `json:"collection,omitempty"`
}

type DeleteCollectionResponse struct{}

func (a *CollectionsAPI) DeleteCollection(ctx context.Context, req *DeleteCollectionRequest) (*DeleteCollectionResponse, error) {
	if err := provided("collection", "is", req.Collection); err != nil {
		return nil, err
	}
	if err := a.store.Delete(ctx, req.Collection); err != nil {
		return nil, err
	}
	return &DeleteCollectionResponse{}, nil
}

type DeleteCollectionsBatchRequest struct {
	Collections []string `json:"collections,omitempty"`
}

type DeleteCollectionsBatchResponse struct{}

func (a *CollectionsAPI) DeleteCollectionsBatch(ctx context.Context, req *DeleteCollectionsBatchRequest) (*DeleteCollectionsBatchResponse, error) {
	if err := provided("collections", "is", req.Collections); err != nil {
		return nil, err
	}
	if err := a.store.DeleteBatch(ctx, req.Collections); err != nil {
		return nil, err
	}
	return &DeleteCollectionsBatchResponse{}, nil
}
