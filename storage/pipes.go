package storage

import (
	"context"
	"database/sql"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/status"
)

type PipeStore struct {
	db *sql.DB
}

func NewPipeStore(db *sql.DB) *PipeStore {
	return &PipeStore{db: db}
}

func (s *PipeStore) Len(ctx context.Context) (int, bool) {
	return 0, false
}

func (s *PipeStore) Delete(ctx context.Context, id string) error {
	return status.ErrUnimplemented
}

func (s *PipeStore) Exists(ctx context.Context, id string) (bool, error) {
	return false, status.ErrUnimplemented
}

func (s *PipeStore) Load(ctx context.Context, id string) (*api.PipeRich, error) {
	return nil, status.ErrUnimplemented
}

func (s *PipeStore) Store(ctx context.Context, object *api.PipeRich) error {
	return status.ErrUnimplemented
}

func (s *PipeStore) StoreOrReplace(ctx context.Context, object *api.PipeRich) error {
	return status.ErrUnimplemented
}

func (s *PipeStore) Scan(ctx context.Context, scanFn func(object *api.PipeRich, err error) error) error {
	return status.ErrUnimplemented
}

func (s *PipeStore) ScanHostTriggered(ctx context.Context, hostname string, scanFn func(*api.PipeRich, error) error) error {
	return status.ErrUnimplemented
}
