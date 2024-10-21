package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"slices"

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

func (s *CollectionStore) LoadBatch(ctx context.Context, ids []string) (results []*api.Collection, missing []string, err error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer conn.Close()

	prep, err := conn.PrepareContext(ctx, fmt.Sprintf("SELECT c.CollectionID, c.Spec FROM Collections AS c WHERE c.CollectionID IN (%s)", placeholders(len(ids))))
	if err != nil {
		return nil, nil, err
	}
	rows, err := prep.QueryContext(ctx, convertToAnySlice(ids)...)
	if err != nil {
		return nil, nil, err
	}
	remaining := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		remaining[id] = struct{}{}
	}
	for rows.Next() {
		var id string
		spec := sql.NullString{}
		if err := rows.Scan(&id, &spec); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue // ID is missing, skip it.
			}
			return nil, nil, err
		}
		delete(remaining, id) // ID is present.
		c := new(api.Collection)
		if err := unmarshalNullableJSONString(spec, c); err != nil {
			return nil, nil, err
		}
		results = append(results, c)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	missing = make([]string, 0, len(remaining))
	for id := range remaining {
		missing = append(missing, id)
	}
	slices.Sort(missing)
	return results, missing, nil
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
	if err := prep.QueryRowContext(ctx, id).Scan(&spec); err != nil {
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
	defer prep.Close()

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

	prepPipes, err := tx.PrepareContext(ctx, fmt.Sprintf("INSERT INTO CollectionPipes (CollectionID, PipeName, PipeDigest) VALUES (%s)", placeholders(3)))
	if err != nil {
		return "", err
	}
	defer prepPipes.Close()

	for _, p := range object.GetPipes() {
		nd, err := api.ParseNameDigest(p)
		if err != nil {
			return "", err
		}
		if !nd.HasDigest() {
			return "", fmt.Errorf("pipe referenced by collection must have a digest (%q): %w", p, status.ErrInvalidArgument)
		}
		if _, err := prepPipes.ExecContext(ctx,
			id,
			nd.GetName(),
			nd.GetDigest(),
		); err != nil {
			return "", err
		}
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
    p.Spec AS PipeSpec
FROM Collections AS c
LEFT JOIN CollectionPipes AS cp USING (CollectionID)
LEFT JOIN Pipes AS p ON cp.PipeName = p.Name AND cp.PipeDigest = p.Digest
WHERE c.Hostname = ?
    OR c.AlwaysTrigger`

func (s *CollectionStore) ScanTriggered(ctx context.Context, u *url.URL, scanFn func(id string, c *api.Collection, pipes []*api.Pipe, err error) error) error {
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
	rows, err := prep.QueryContext(ctx, hostname)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		spec := sql.NullString{}
		pipeSpec := sql.NullString{}
		if err := rows.Scan(&id, &spec, &pipeSpec); err != nil {
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
		pipe := new(api.Pipe)
		if err := unmarshalNullableJSONString(pipeSpec, pipe); err != nil {
			return err
		}
		if err := scanFn(id, c, []*api.Pipe{pipe}, nil); err != nil {
			if err == ErrStopScan {
				break
			}
			return err
		}
	}
	return rows.Err()
}
