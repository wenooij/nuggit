package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/status"
)

type CollectionStore struct{ db *sql.DB }

func NewCollectionStore(db *sql.DB) *CollectionStore {
	return &CollectionStore{db: db}
}

func (s *CollectionStore) Len(ctx context.Context) (int, bool) {
	return tableCount(ctx, s.db, "Collections")
}

func (s *CollectionStore) Delete(ctx context.Context, id string) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	prep, err := tx.PrepareContext(ctx, "DELETE FROM Collections WHERE CollectionID = ?")
	if err != nil {
		return err
	}
	if _, err := prep.ExecContext(ctx, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *CollectionStore) LoadBatch(ctx context.Context, ids []string) ([]*api.Collection, error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	prep, err := conn.PrepareContext(ctx, fmt.Sprintf("SELECT c.CollectionID, c.Spec, c.Conditions FROM Collections AS c WHERE CollectionID IN (%s)", placeholders(len(ids))))
	if err != nil {
		return nil, err
	}
	rows, err := prep.QueryContext(ctx, convertToAnySlice(ids)...)
	if err != nil {
		return nil, err
	}
	var results []*api.Collection
	for rows.Next() {
		var id string
		spec := sql.NullString{}
		conditions := sql.NullString{}
		if err := rows.Scan(&id, &spec, &conditions); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, status.ErrNotFound
			}
			return nil, err
		}
		c := &api.Collection{
			CollectionLite: api.NewCollectionLite(id),
			CollectionBase: new(api.CollectionBase),
			Conditions:     new(api.CollectionConditions),
		}
		if err := unmarshalNullableJSONString(spec, c.CollectionBase); err != nil {
			return nil, err
		}
		if err := unmarshalNullableJSONString(conditions, c.Conditions); err != nil {
			return nil, err
		}
		results = append(results, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (s *CollectionStore) DeleteBatch(ctx context.Context, ids []string) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	prep, err := tx.PrepareContext(ctx, fmt.Sprintf("DELETE FROM Collections WHERE CollectionID IN (%s)", placeholders(len(ids))))
	if err != nil {
		return err
	}
	if _, err := prep.ExecContext(ctx, convertToAnySlice(ids)...); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *CollectionStore) Exists(ctx context.Context, id string) (bool, error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	var i int64
	if err := conn.QueryRowContext(ctx, "SELECT 1 FROM Collections AS c WHERE c.CollectionID = ? LIMIT 1", id).Scan(&i); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *CollectionStore) Load(ctx context.Context, id string) (*api.Collection, error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	prep, err := conn.PrepareContext(ctx, "SELECT c.Spec, c.Conditions FROM Collections AS c WHERE c.CollectionID = ? LIMIT 1")
	if err != nil {
		return nil, err
	}
	spec := sql.NullString{}
	conditions := sql.NullString{}
	if err := prep.QueryRowContext(ctx, id).Scan(&spec, &conditions); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.ErrNotFound
		}
		return nil, err
	}
	c := &api.Collection{
		CollectionLite: api.NewCollectionLite(id),
		CollectionBase: new(api.CollectionBase),
		Conditions:     new(api.CollectionConditions),
	}
	if err := unmarshalNullableJSONString(spec, c.CollectionBase); err != nil {
		return nil, err
	}
	if err := unmarshalNullableJSONString(conditions, c.Conditions); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CollectionStore) Store(ctx context.Context, object *api.Collection) (err error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	id := object.UUID()
	prep, err := tx.PrepareContext(ctx, "SELECT 1 FROM Collections WHERE CollectionID = ?")
	if err != nil {
		return err
	}
	var i int64
	if err := prep.QueryRowContext(ctx, id).Scan(&i); err == nil {
		return status.ErrAlreadyExists
	} else if errors.Is(err, sql.ErrNoRows) {
		return s.storeOrReplaceCollectionTx(ctx, tx, object)
	} else {
		return err
	}
}

func (s *CollectionStore) StoreOrReplace(ctx context.Context, object *api.Collection) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	return s.storeOrReplaceCollectionTx(ctx, tx, object)
}

func (s *CollectionStore) storeOrReplaceCollectionTx(ctx context.Context, tx *sql.Tx, object *api.Collection) error {
	id := object.UUID()

	spec, err := marshalNullableJSONString(object.GetBase())
	if err != nil {
		return err
	}

	conditions, err := marshalNullableJSONString(object.GetConditions())
	if err != nil {
		return err
	}

	prep, err := tx.PrepareContext(ctx, "INSERT INTO Collections (CollectionID, Name, AlwaysTrigger, Hostname, URLPattern, Spec, Conditions) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}

	if _, err := prep.ExecContext(ctx,
		id,
		object.GetBase().GetName(),
		object.GetConditions().GetAlwaysTrigger(),
		object.GetConditions().GetHostname(),
		object.GetConditions().GetURLPattern(),
		spec,
		conditions,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *CollectionStore) Scan(ctx context.Context, scanFn func(*api.CollectionLite, error) error) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	rows, err := conn.QueryContext(ctx, "SELECT CollectionID FROM Collections")
	if err != nil {
		return err
	}
	for rows.Next() {
		var id sql.NullString
		err := rows.Scan(&id)
		if err := scanFn(api.NewCollectionLite(id.String), err); err != nil {
			if err == ErrStopScan {
				return nil
			}
			return err
		}
	}
	return rows.Err()
}

func (s *CollectionStore) ScanTriggered(ctx context.Context, u *url.URL, scanFn func(*api.Collection, error) error) error {
	return status.ErrUnimplemented
}
