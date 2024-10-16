package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"regexp"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/status"
)

type CollectionStore struct{ db *sql.DB }

func NewCollectionStore(db *sql.DB) *CollectionStore {
	return &CollectionStore{db: db}
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

	prep, err := conn.PrepareContext(ctx, fmt.Sprintf("SELECT c.CollectionID, c.Spec FROM Collections AS c WHERE CollectionID IN (%s)", placeholders(len(ids))))
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
		if err := rows.Scan(&id, &spec); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, status.ErrNotFound
			}
			return nil, err
		}
		c := new(api.Collection)
		if err := unmarshalNullableJSONString(spec, c); err != nil {
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

	prep, err := conn.PrepareContext(ctx, "SELECT c.Spec FROM Collections AS c WHERE c.CollectionID = ? LIMIT 1")
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
	c := new(api.Collection)
	if err := unmarshalNullableJSONString(spec, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CollectionStore) Lookup(ctx context.Context, name string) (string, error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	prep, err := conn.PrepareContext(ctx, "SELECT c.CollectionID FROM Collections AS c WHERE c.Name = ? LIMIT 1")
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

func (s *CollectionStore) LoadName(ctx context.Context, name string) (string, *api.Collection, error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return "", nil, err
	}
	defer conn.Close()

	prep, err := conn.PrepareContext(ctx, "SELECT c.CollectionID, c.Spec FROM Collections AS c WHERE c.Name = ? LIMIT 1")
	if err != nil {
		return "", nil, err
	}
	var id string
	spec := sql.NullString{}
	if err := prep.QueryRowContext(ctx, id).Scan(&id, &spec); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil, status.ErrNotFound
		}
		return "", nil, err
	}
	c := new(api.Collection)
	if err := unmarshalNullableJSONString(spec, c); err != nil {
		return "", nil, err
	}
	return id, c, nil
}

func (s *CollectionStore) Store(ctx context.Context, object *api.Collection) (string, error) {
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

	prep, err := tx.PrepareContext(ctx, "INSERT INTO Collections (CollectionID, Name, AlwaysTrigger, Hostname, URLPattern, Spec) VALUES (?, ?, ?, ?, ?, ?)")
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
		object.GetConditions().GetAlwaysTrigger(),
		object.GetConditions().GetHostname(),
		object.GetConditions().GetURLPattern(),
		spec,
	); err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return id, nil
}

func (s *CollectionStore) ScanRef(ctx context.Context, scanFn func(string, error) error) error {
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

const triggerQuery = `SELECT
	c.CollectionID,
	c.Spec,
	c.Conditions
FROM Collections AS c
WHERE c.Hostname = ?
	OR URLPattern IS NOT NULL AND URLPattern != ''
	OR AlwaysTrigger`

func (s *CollectionStore) ScanTriggered(ctx context.Context, u *url.URL, scanFn func(string, *api.Collection, error) error) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	prep, err := conn.PrepareContext(ctx, triggerQuery)
	if err != nil {
		return err
	}
	urlStr := u.String()
	hostname := u.Hostname()
	rows, err := prep.QueryContext(ctx, triggerQuery, hostname)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id string
		spec := sql.NullString{}
		if err := rows.Scan(&id, &spec); err != nil {
			return err
		}
		c := new(api.Collection)
		if err := unmarshalNullableJSONString(spec, c); err != nil {
			return err
		}
		if c.GetConditions() == nil {
			continue // No trigger.
		}
		trigger := c.GetConditions().GetAlwaysTrigger() || c.GetConditions().GetHostname() == hostname
		if pattern := c.GetConditions().GetURLPattern(); !trigger && pattern != "" {
			match, err := regexp.MatchString(pattern, urlStr)
			if err != nil {
				return err
			}
			trigger = match
		}
		if !trigger {
			continue // No trigger.
		}
		if err := scanFn(id, c, nil); err != nil {
			if err == ErrStopScan {
				break
			}
			return err
		}
	}
	return rows.Err()
}
