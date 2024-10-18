package storage

import (
	"context"
	"database/sql"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/status"
)

type TriggerStore struct {
	db *sql.DB
}

func NewTriggerStore(db *sql.DB) *TriggerStore {
	return &TriggerStore{db: db}
}

func (s *TriggerStore) Delete(ctx context.Context, id string) error {
	return status.ErrUnimplemented
}

func (s *TriggerStore) Load(ctx context.Context, id string) (*api.TriggerRecord, error) {
	return nil, status.ErrUnimplemented
}

func (s *TriggerStore) Store(ctx context.Context, object *api.TriggerRecord) (string, error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	spec, err := marshalNullableJSONString(object.GetTrigger())
	if err != nil {
		return "", err
	}

	plan, err := marshalNullableJSONString(object.GetPlan())
	if err != nil {
		return "", err
	}

	prep, err := tx.PrepareContext(ctx, "INSERT INTO Triggers (TriggerID, Committed, Plan, Spec) VALUES (?, false, ?, ?)")
	if err != nil {
		return "", err
	}

	id, err := newUUID()
	if err != nil {
		return "", err
	}

	if _, err := prep.ExecContext(ctx,
		id,
		spec,
		plan,
	); err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return id, nil
}

func (s *TriggerStore) Scan(ctx context.Context, scanFn func(object *api.Trigger, err error) error) error {
	return status.ErrUnimplemented
}
