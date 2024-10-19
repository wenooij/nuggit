package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/status"
)

type PipeStore struct {
	db *sql.DB
}

func NewPipeStore(db *sql.DB) *PipeStore {
	return &PipeStore{db: db}
}

func (s *PipeStore) Delete(ctx context.Context, pipeDigest string) error {
	return status.ErrUnimplemented
}

func (s *PipeStore) DeleteBatch(ctx context.Context, pipeDigests []string) error {
	return status.ErrUnimplemented
}

func (s *PipeStore) Load(ctx context.Context, pipeDigest string) (*api.Pipe, error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	name, digest, err := api.SplitPipeDigest(pipeDigest)
	if err != nil {
		return nil, err
	}

	prep, err := conn.PrepareContext(ctx, "SELECT p.Spec FROM Pipes AS p WHERE p.Name = ? AND p.Digest = ? LIMIT 1")
	if err != nil {
		return nil, err
	}
	spec := sql.NullString{}
	if err := prep.QueryRowContext(ctx, name, digest).Scan(&spec); err != nil {
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

func (s *PipeStore) LoadBatch(ctx context.Context, pipeDigests []string) (results []*api.Pipe, missing []string, err error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer conn.Close()

	prep, err := conn.PrepareContext(ctx, fmt.Sprintf(`SELECT
	CONCAT(p.Name, '@', p.Digest) AS pipeDigest,
	p.Spec
FROM Pipes AS p
WHERE CONCAT(p.Name, '@', p.Digest) IN (%s)`, placeholders(len(pipeDigests))))
	if err != nil {
		return nil, nil, err
	}
	rows, err := prep.QueryContext(ctx, convertToAnySlice(pipeDigests)...)
	if err != nil {
		return nil, nil, err
	}
	remaining := make(map[string]struct{}, len(pipeDigests))
	for _, pipeDigest := range pipeDigests {
		remaining[pipeDigest] = struct{}{}
	}
	for rows.Next() {
		var pipeDigest string
		spec := sql.NullString{}
		if err := rows.Scan(&pipeDigest, &spec); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue // ID is missing, skip it.
			}
			return nil, nil, err
		}
		delete(remaining, pipeDigest) // ID is present.
		p := new(api.Pipe)
		if err := unmarshalNullableJSONString(spec, p); err != nil {
			return nil, nil, err
		}
		results = append(results, p)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	missing = make([]string, 0, len(remaining))
	for pipeDigest := range remaining {
		missing = append(missing, pipeDigest)
	}
	slices.Sort(missing)
	return results, missing, nil
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

	prep, err := tx.PrepareContext(ctx, "INSERT INTO Pipes (Name, Digest, Spec) VALUES (?, ?, ?)")
	if err != nil {
		return "", err
	}

	name := object.GetName()
	digest, err := api.PipeDigestSHA1(object)
	if err != nil {
		return "", err
	}
	pipeDigest, err := api.JoinPipeDigest(name, digest)
	if err != nil {
		return "", err
	}

	if _, err := prep.ExecContext(ctx,
		name,
		digest,
		spec,
	); err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return pipeDigest, nil
}

func (s *PipeStore) Scan(ctx context.Context, scanFn func(object *api.Pipe, err error) error) error {
	return status.ErrUnimplemented
}

func (s *PipeStore) ScanRef(ctx context.Context, scanFn func(pipeDigest string, err error) error) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	rows, err := conn.QueryContext(ctx, "SELECT CONCAT(p.Name, '@', p.Digest) AS pipeDigest FROM Pipes AS p")
	if err != nil {
		return err
	}
	for rows.Next() {
		var pipeDigest string
		err := rows.Scan(&pipeDigest)
		if err := scanFn(pipeDigest, err); err != nil {
			if err == ErrStopScan {
				return nil
			}
			return err
		}
	}
	return rows.Err()
}
