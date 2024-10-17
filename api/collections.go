package api

import (
	"context"
	"time"
)

type CollectionLite struct {
	*Ref `json:",omitempty"`
}

func NewCollectionLite(id string) *CollectionLite {
	return &CollectionLite{newRef("/api/collections/", id)}
}

func (c *CollectionLite) GetRef() *Ref {
	if c == nil {
		return nil
	}
	return c.Ref
}

type CollectionBase struct {
	Name            string   `json:"name,omitempty"`
	Points          []*Point `json:"row,omitempty"`
	IncludeMetadata bool     `json:"include_metadata,omitempty"`
}

func (c *CollectionBase) GetName() string {
	if c == nil {
		return ""
	}
	return c.Name
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
	*CollectionLite `json:",omitempty"`
	*CollectionBase `json:",omitempty"`
	Conditions      *CollectionConditions `json:"conditions,omitempty"`
	State           *CollectionState      `json:"state,omitempty"`
}

func (c *Collection) GetLite() *CollectionLite {
	if c == nil {
		return nil
	}
	return c.CollectionLite
}

func (c *Collection) GetBase() *CollectionBase {
	if c == nil {
		return nil
	}
	return c.CollectionBase
}

func (c *Collection) GetConditions() *CollectionConditions {
	if c == nil {
		return nil
	}
	return c.Conditions
}

func (c *Collection) GetState() *CollectionState {
	if c == nil {
		return nil
	}
	return c.State
}

type CollectionState struct {
	Pipes map[string]struct{} `json:"pipes,omitempty"`
}

func (s *CollectionState) GetPipes() map[string]struct{} {
	if s == nil {
		return nil
	}
	return s.Pipes
}

type CollectionDataBase struct {
	*CollectionLite
	*PointValues `json:",omitempty"`
}

type CollectionData struct {
	*CollectionLite
}

type PointMetadata struct {
	URL                 string    `json:"url,omitempty"`
	Timestamp           time.Time `json:"timestamp,omitempty"`
	PageContentChecksum string    `json:"page_content_checksum,omitempty"`
}

type PointValues struct {
	Metadata *PointMetadata `json:"metadata,omitempty"`
	Values   []any          `json:"values,omitempty"`
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
	Collection *CollectionBase       `json:"collection,omitempty"`
	Conditions *CollectionConditions `json:"conditions,omitempty"`
}

type CreateCollectionResponse struct {
	Collection *CollectionLite `json:"collection,omitempty"`
}

func (a *CollectionsAPI) CreateCollection(ctx context.Context, req *CreateCollectionRequest) (*CreateCollectionResponse, error) {
	if err := provided("collection", "is", req.Collection); err != nil {
		return nil, err
	}
	if err := provided("name", "is", req.Collection.Name); err != nil {
		return nil, err
	}
	id, err := newUUID(ctx, a.store.Exists)
	if err != nil {
		return nil, err
	}
	cl := NewCollectionLite(id)
	if err := a.store.StoreOrReplace(ctx, &Collection{
		CollectionLite: cl,
		CollectionBase: req.Collection,
		Conditions:     req.Conditions,
	}); err != nil {
		return nil, err
	}
	return &CreateCollectionResponse{
		Collection: cl,
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
	Collections []*CollectionLite `json:"collections,omitempty"`
}

func (a *CollectionsAPI) ListCollections(ctx context.Context, req *ListCollectionsRequest) (*ListCollectionsResponse, error) {
	var res []*CollectionLite
	if err := a.store.Scan(ctx, func(cl *CollectionLite, err error) error {
		if err != nil {
			return err
		}
		res = append(res, cl)
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
