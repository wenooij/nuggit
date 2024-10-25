package storage

import (
	"context"
	"database/sql"
	"fmt"
	"iter"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/status"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

type PipeStore struct {
	db *sql.DB
}

func NewPipeStore(db *sql.DB) *PipeStore {
	return &PipeStore{db: db}
}

func (s *PipeStore) Delete(ctx context.Context, name api.NameDigest) error {
	return deleteSpec(ctx, s.db, "Pipes", name)
}

func (s *PipeStore) DeleteBatch(ctx context.Context, names []api.NameDigest) error {
	return deleteBatch(ctx, s.db, "Pipes", names)
}

func (s *PipeStore) Load(ctx context.Context, name api.NameDigest) (*api.Pipe, error) {
	c := new(api.Pipe)
	if err := loadSpec(ctx, s.db, "Pipes", name, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *PipeStore) LoadBatch(ctx context.Context, names []api.NameDigest) iter.Seq2[*api.Pipe, error] {
	return scanSpecsBatch(ctx, s.db, "Pipes", names, func() *api.Pipe { return new(api.Pipe) })
}

func (s *PipeStore) Store(ctx context.Context, object *api.Pipe) error {
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

	spec, err := marshalNullableJSONString(object)
	if err != nil {
		return err
	}

	prep, err := tx.PrepareContext(ctx, "INSERT INTO Pipes (Name, Digest, Spec) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer prep.Close()

	nd, err := api.NewNameDigest(object)
	if err != nil {
		return err
	}
	if _, err := prep.ExecContext(ctx,
		nd.GetName(),
		nd.GetDigest(),
		spec,
	); err != nil {
		if err, ok := err.(*sqlite.Error); ok {
			if err.Code() == sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY {
				return fmt.Errorf("failed to store pipe: %w", status.ErrAlreadyExists)
			}
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *PipeStore) StoreBatch(ctx context.Context, objects []*api.Pipe) error {
	// FIXME: Use a shared db connection for StoreBatch and Store.
	for _, o := range objects {
		err := s.Store(ctx, o)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PipeStore) Scan(ctx context.Context) iter.Seq2[*api.Pipe, error] {
	return scanSpecs(ctx, s.db, "Pipes", func() *api.Pipe { return new(api.Pipe) })
}

func (s *PipeStore) ScanNames(ctx context.Context) iter.Seq2[api.NameDigest, error] {
	return scanNames(ctx, s.db, "Pipes")
}

func (s *PipeStore) StorePipeDependencies(ctx context.Context, pipe api.NameDigest, references []api.NameDigest) error {
	if len(references) == 0 {
		return nil
	}

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

	prep, err := tx.PrepareContext(ctx, "INSERT INTO PipeReferences (Name, Digest, ReferencedName, ReferencedDigest) VALUES (?, ?, ?, ?) ON CONFLICT DO NOTHING")
	if err != nil {
		return err
	}
	defer prep.Close()

	for _, r := range references {
		if _, err := prep.ExecContext(ctx,
			pipe.GetName(),
			pipe.GetDigest(),
			r.GetName(),
			r.GetDigest(),
		); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *PipeStore) ScanDependencies(ctx context.Context, pipe api.NameDigest) iter.Seq2[*api.Pipe, error] {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return seq2Error[*api.Pipe](err)
	}

	prep, err := conn.PrepareContext(ctx, `SELECT
	pp.ReferencedName AS Name,
	pp.ReferencedDigest AS Digest,
	p.Spec
FROM PipeReferences AS pp
JOIN Pipes AS p ON pp.ReferencedName = p.Name AND pp.ReferencedDigest = p.Digest
WHERE pp.Name = ? AND pp.Digest = ?`)
	if err != nil {
		return seq2Error[*api.Pipe](err)
	}

	rows, err := prep.QueryContext(ctx, pipe.GetName(), pipe.GetDigest())
	if err != nil {
		return seq2Error[*api.Pipe](err)
	}

	return func(yield func(*api.Pipe, error) bool) {
		defer conn.Close()
		defer prep.Close()
		defer rows.Close()

		for rows.Next() {
			var name, digest, spec sql.NullString
			if err := rows.Scan(&name, &digest, &spec); err != nil {
				yield(nil, err)
				return
			}
			p := new(api.Pipe)
			if err := unmarshalNullableJSONString(spec, p); err != nil {
				yield(nil, err)
				return
			}
			p.SetNameDigest(api.NameDigest{Name: name.String, Digest: digest.String})
			if !yield(p, nil) {
				break
			}
		}
		if err := rows.Err(); err != nil {
			yield(nil, err)
		}
	}
}
