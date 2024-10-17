package storage

import (
	"context"
	"database/sql"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/status"
)

type NodeStore struct {
	db *sql.DB
}

func NewNodeStore(db *sql.DB) *NodeStore {
	return &NodeStore{db: db}
}

func (s *NodeStore) Len(ctx context.Context) (int, bool) {
	return 0, false
}

func (s *NodeStore) Delete(ctx context.Context, id string) error {
	return status.ErrUnimplemented
}

func (s *NodeStore) Load(ctx context.Context, id string) (*api.NodeRich, error) {
	return nil, status.ErrUnimplemented
}

func (s *NodeStore) Store(ctx context.Context, object *api.NodeRich) error {
	return status.ErrUnimplemented
}

func (s *NodeStore) StoreOrReplace(ctx context.Context, object *api.NodeRich) error {
	return status.ErrUnimplemented
}

func (s *NodeStore) Scan(ctx context.Context, scanFn func(object *api.NodeRich, err error) error) error {
	return status.ErrUnimplemented
}
