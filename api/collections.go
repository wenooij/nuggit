package api

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/wenooij/nuggit/status"
)

type CollectionLite struct {
	*Ref `json:",omitempty"`
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
	storage StoreInterface[*Collection]
	mu      sync.RWMutex
}

func (a *CollectionsAPI) Init(api *API, storeType StorageType) error {
	*a = CollectionsAPI{
		api: api,
	}
	if storeType != StorageInMemory {
		return fmt.Errorf("persistent collections not supported: %w", status.ErrUnimplemented)
	}
	a.storage = newStorageInMemory[*Collection]()
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
	a.mu.Lock()
	defer a.mu.Unlock()
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
	cl := &CollectionLite{&Ref{
		ID:  id,
		URI: fmt.Sprintf("/api/collections/%s", id),
	}}
	storeOp, err := a.storage.Store(&Collection{
		CollectionLite: cl,
		CollectionBase: req.Collection,
	})
	if err != nil {
		return nil, err
	}
	return &CreateCollectionResponse{
		StoreOp:    storeOp,
		Collection: cl,
	}, nil
}

type GetCollectionRequest struct {
	Collection *CollectionLite `json:"collection,omitempty"`
}

type GetCollectionResponse struct {
	Collection *Collection `json:"collection,omitempty"`
}

func (a *CollectionsAPI) GetCollection(req *GetCollectionRequest) (*GetCollectionResponse, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if err := provided("collection", "is", req.Collection); err != nil {
		return nil, err
	}
	id := req.Collection.UUID()
	collection, err := a.storage.Load(id)
	if err != nil {
		return nil, err
	}
	return &GetCollectionResponse{Collection: collection}, nil
}
