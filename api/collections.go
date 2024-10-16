package api

import (
	"errors"
	"fmt"
	"time"

	"github.com/wenooij/nuggit/status"
)

type CollectionLite struct {
	*Ref `json:",omitempty"`
}

func newCollectionLite(id string) *CollectionLite {
	return &CollectionLite{&Ref{
		ID:  id,
		URI: fmt.Sprintf("/api/collections/%s", id),
	}}
}

type CollectionBase struct {
	Name            string   `json:"name,omitempty"`
	Points          []*Point `json:"row,omitempty"`
	DryRun          bool     `json:"dry_run,omitempty"`
	IncludeMetadata bool     `json:"include_metadata,omitempty"`
}

type Collection struct {
	*CollectionLite `json:",omitempty"`
	*CollectionBase `json:",omitempty"`
	*PointValues    `json:",omitempty"`
}

type CollectionState struct {
	Pipelines map[string]struct{} `json:"pipelines,omitempty"`
}

type CollectionRich struct {
	*Collection `json:",omitempty"`
	State       *CollectionState `json:"state,omitempty"`
}

type Point struct {
	Pipe *PipeLite `json:"pipe,omitempty"`
	Name string    `json:"name,omitempty"`
	Type Type      `json:"type,omitempty"`
}

func (p *Point) GetType() Type {
	if p == nil {
		return scalar(TypeUndefined)
	}
	return p.Type
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
	api     *API
	storage StoreInterface[*CollectionRich]
}

func (a *CollectionsAPI) Init(storeType StorageType) error {
	*a = CollectionsAPI{}
	if storeType != StorageInMemory {
		return fmt.Errorf("persistent collections not supported: %w", status.ErrUnimplemented)
	}
	a.storage = newStorageInMemory[*CollectionRich]()
	return nil
}

type CreateCollectionRequest struct {
	Collection *CollectionBase `json:"collection,omitempty"`
}

type CreateCollectionResponse struct {
	StoreOp    *StorageOpLite  `json:"store_op,omitempty"`
	Collection *CollectionLite `json:"collection,omitempty"`
}

func (a *CollectionsAPI) CreateCollection(req *CreateCollectionRequest) (*CreateCollectionResponse, error) {
	if err := provided("collection", "is", req.Collection); err != nil {
		return nil, err
	}
	if err := provided("name", "is", req.Collection.Name); err != nil {
		return nil, err
	}
	id, err := newUUID(func(id string) bool { _, err := a.storage.Load(id); return errors.Is(err, status.ErrNotFound) })
	if err != nil {
		return nil, err
	}
	cl := newCollectionLite(id)
	storeOp, err := a.storage.Store(&CollectionRich{Collection: &Collection{
		CollectionLite: cl,
		CollectionBase: req.Collection,
	}})
	if err != nil {
		return nil, err
	}
	return &CreateCollectionResponse{
		StoreOp:    storeOp,
		Collection: cl,
	}, nil
}

type GetCollectionRequest struct {
	Collection string `json:"collection,omitempty"`
}

type GetCollectionResponse struct {
	Collection *CollectionRich `json:"collection,omitempty"`
}

func (a *CollectionsAPI) GetCollection(req *GetCollectionRequest) (*GetCollectionResponse, error) {
	if err := provided("collection", "is", req.Collection); err != nil {
		return nil, err
	}
	collection, err := a.storage.Load(req.Collection)
	if err != nil {
		return nil, err
	}
	return &GetCollectionResponse{Collection: collection}, nil
}
