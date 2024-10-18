package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/status"
)

type TriggerStore struct {
	db *sql.DB
}

func NewTriggerStore(db *sql.DB) *TriggerStore {
	return &TriggerStore{db: db}
}

func (s *TriggerStore) Len(ctx context.Context) (int, bool) {
	return 0, false
}

func (s *TriggerStore) Delete(ctx context.Context, id string) error {
	return status.ErrUnimplemented
}

func (s *TriggerStore) Exists(ctx context.Context, id string) (bool, error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	var i int64
	if err := conn.QueryRowContext(ctx, "SELECT 1 FROM Triggers AS t WHERE t.TriggerID = ? LIMIT 1", id).Scan(&i); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *TriggerStore) Load(ctx context.Context, id string) (*api.Trigger, error) {
	return nil, status.ErrUnimplemented
}

func (s *TriggerStore) Store(ctx context.Context, object *api.Trigger) error {
	return status.ErrUnimplemented
}

func (s *TriggerStore) StoreOrReplace(ctx context.Context, object *api.Trigger) error {
	return status.ErrUnimplemented
}

func (s *TriggerStore) Scan(ctx context.Context, scanFn func(object *api.Trigger, err error) error) error {
	return status.ErrUnimplemented
}
