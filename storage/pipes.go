package storage

import (
	"context"
	"database/sql"
	"errors"

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

func (s *PipeStore) DeleteBatch(ctx context.Context, ids []string) error {
	return status.ErrUnimplemented
}

func (s *PipeStore) Exists(ctx context.Context, id string) (bool, error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	var i int64
	if err := conn.QueryRowContext(ctx, "SELECT 1 FROM Pipes AS p WHERE p.PipeID = ? LIMIT 1", id).Scan(&i); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *PipeStore) Load(ctx context.Context, id string) (*api.Pipe, error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	prep, err := conn.PrepareContext(ctx, "SELECT p.Spec FROM Pipes AS p WHERE p.PipeID = ? LIMIT 1")
	if err != nil {
		return nil, err
	}
	spec := sql.NullString{}
	if err := prep.QueryRowContext(ctx, id).Scan(&spec); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.ErrNotFound
		}
		return nil, err
	}
	p := new(api.Pipe)
	if err := unmarshalNullableJSONString(spec, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PipeStore) Lookup(ctx context.Context, name string) (string, error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	prep, err := conn.PrepareContext(ctx, "	")
	if err != nil {
		return "", err
	}
	var id string
	if err := prep.QueryRowContext(ctx, id).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", status.ErrNotFound
		}
		return "", err
	}
	return id, nil
}

func (s *PipeStore) LoadBatch(ctx context.Context, ids []string) ([]*api.Pipe, error) {
	return nil, status.ErrUnimplemented
}

func (s *PipeStore) Store(ctx context.Context, object *api.Pipe) (string, error) {
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

	spec, err := marshalNullableJSONString(object)
	if err != nil {
		return "", err
	}

	prep, err := tx.PrepareContext(ctx, "INSERT INTO Pipes (PipeID, Name, Spec) VALUES (?, ?, ?)")
	if err != nil {
		return "", err
	}

	id, err := newUUID()
	if err != nil {
		return "", err
	}

	if _, err := prep.ExecContext(ctx,
		id,
		object.GetName(),
		spec,
	); err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return id, nil
}

func (s *PipeStore) Scan(ctx context.Context, scanFn func(object *api.Pipe, err error) error) error {
	return status.ErrUnimplemented
}

func (s *PipeStore) ScanRef(ctx context.Context, scanFn func(id string, err error) error) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	rows, err := conn.QueryContext(ctx, "SELECT p.PipeID FROM Pipes AS p")
	if err != nil {
		return err
	}
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err := scanFn(id, err); err != nil {
			if err == ErrStopScan {
				return nil
			}
			return err
		}
	}
	return rows.Err()
}
