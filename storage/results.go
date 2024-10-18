package storage

import (
	"context"
	"database/sql"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/status"
)

type TriggerResultStore struct {
	db *sql.DB
}

func NewTriggerResultStore(db *sql.DB) *TriggerResultStore {
	return &TriggerResultStore{db: db}
}

func (s *TriggerResultStore) Len(ctx context.Context) (int, bool) {
	return 0, false
}

func (s *TriggerResultStore) Delete(ctx context.Context, id string) error {
	return status.ErrUnimplemented
}

func (s *TriggerResultStore) Exists(ctx context.Context, id string) (bool, error) {
	return false, status.ErrUnimplemented
}

func (s *TriggerResultStore) Load(ctx context.Context, id string) (*api.TriggerResult, error) {
	return nil, status.ErrUnimplemented
}

func (s *TriggerResultStore) Store(ctx context.Context, object *api.TriggerResult) error {
	return status.ErrUnimplemented
}

func (s *TriggerResultStore) StoreOrReplace(ctx context.Context, object *api.TriggerResult) error {
	return status.ErrUnimplemented
}

func (s *TriggerResultStore) Scan(ctx context.Context, scanFn func(object *api.TriggerResult, err error) error) error {
	return status.ErrUnimplemented
}
