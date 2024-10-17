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

type CollectionBase struct {
	Name            string                `json:"name,omitempty"`
	Points          []*Point              `json:"row,omitempty"`
	DryRun          bool                  `json:"dry_run,omitempty"`
	IncludeMetadata bool                  `json:"include_metadata,omitempty"`
	Conditions      *CollectionConditions `json:"conditions,omitempty"`
}

type Point struct {
	Name string `json:"name,omitempty"`
	Type Type   `json:"type,omitempty"`
}

func (p *Point) GetType() Type {
	if p == nil {
		return scalar(TypeUndefined)
	}
	return p.Type
}

type CollectionConditions struct {
	AlwaysTrigger bool   `json:"always_trigger,omitempty"`
	Host          string `json:"host,omitempty"`
	URLPattern    string `json:"url_pattern,omitempty"`
}

type Collection struct {
	*CollectionLite `json:",omitempty"`
	*CollectionBase `json:",omitempty"`
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

type CollectionRich struct {
	*Collection `json:",omitempty"`
	State       *CollectionState `json:"state,omitempty"`
}

func (c *CollectionRich) GetCollection() *Collection {
	if c == nil {
		return nil
	}
	return c.Collection
}

func (c *CollectionRich) GetState() *CollectionState {
	if c == nil {
		return nil
	}
	return c.State
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

func (a *CollectionsAPI) Init(store CollectionStore) error {
	*a = CollectionsAPI{
		store: store,
	}
	return nil
}

type CreateCollectionRequest struct {
	Collection *CollectionBase `json:"collection,omitempty"`
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
	id, err := newUUID(func(id string) error { _, err := a.store.Load(ctx, id); return err })
	if err != nil {
		return nil, err
	}
	cl := NewCollectionLite(id)
	if err := a.store.Store(ctx, &CollectionRich{Collection: &Collection{
		CollectionLite: cl,
		CollectionBase: req.Collection,
	}}); err != nil {
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
	Collection *CollectionRich `json:"collection,omitempty"`
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
