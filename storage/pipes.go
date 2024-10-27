package storage

import (
	"context"
	"database/sql"
	"fmt"
	"iter"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/integrity"
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

func (s *PipeStore) Delete(ctx context.Context, name integrity.NameDigest) error {
	return deleteSpec(ctx, s.db, "Pipes", name)
}

func (s *PipeStore) DeleteBatch(ctx context.Context, names []integrity.NameDigest) error {
	return deleteBatch(ctx, s.db, "Pipes", names)
}

func (s *PipeStore) Load(ctx context.Context, name integrity.NameDigest) (*api.Pipe, error) {
	c := new(api.Pipe)
	if err := loadSpec(ctx, s.db, "Pipes", name, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *PipeStore) LoadBatch(ctx context.Context, names []integrity.NameDigest) iter.Seq2[*api.Pipe, error] {
	return scanSpecsBatch(ctx, s.db, "Pipes", names, func() *api.Pipe { return new(api.Pipe) })
}

func (s *PipeStore) Store(ctx context.Context, pipe *api.Pipe) error {
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

	spec, err := marshalNullableJSONString(pipe)
	if err != nil {
		return err
	}

	prep, err := tx.PrepareContext(ctx, "INSERT INTO Pipes (Name, Digest, TypeNumber, Spec) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer prep.Close()

	nd, err := integrity.NewNameDigest(pipe)
	if err != nil {
		return err
	}
	if _, err := prep.ExecContext(ctx,
		nd.GetName(),
		nd.GetDigest(),
		pipe.GetPoint().AsNumber(),
		spec,
	); err != nil {
		if err, ok := err.(*sqlite.Error); ok {
			if err.Code() == sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY {
				return fmt.Errorf("failed to store pipe: %w", status.ErrAlreadyExists)
			}
		}
		return err
	}

	// Store pipe dependencies.
	var dependencies []integrity.NameDigest
	for _, a := range pipe.GetActions() {
		if a.GetAction() == "pipe" {
			dependencies = append(dependencies, integrity.GetNameDigestArg(a))
		}
	}

	prepDeps, err := tx.PrepareContext(ctx, `INSERT INTO PipeDependencies (PipeID, ReferencedID)
	SELECT p.ID AS PipeID, p2.ID AS ReferencedID
	FROM Pipes AS p
	JOIN Pipes AS p2 ON 1
	WHERE p.Name = ? AND p.Digest = ? AND
		  p2.Name = ? AND p2.Digest = ? LIMIT 1`)
	if err != nil {
		return err
	}
	defer prepDeps.Close()

	for _, dep := range dependencies {
		if _, err := prepDeps.ExecContext(ctx,
			nd.GetName(),
			nd.GetDigest(),
			dep.GetName(),
			dep.GetDigest(),
		); err != nil {
			return err
		}
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

func (s *PipeStore) ScanNames(ctx context.Context) iter.Seq2[integrity.NameDigest, error] {
	return scanNames(ctx, s.db, "Pipes")
}

func (s *PipeStore) ScanDependencies(ctx context.Context, pipe integrity.NameDigest) iter.Seq2[*api.Pipe, error] {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return seq2Error[*api.Pipe](err)
	}

	prep, err := conn.PrepareContext(ctx, `SELECT
	p.Name AS ReferencedName,
	p.Digest AS ReferencedDigest,
	p.Spec
FROM PipeDependencies AS d
JOIN Pipes AS p ON d.PipeID = p.ID
WHERE p.Name = ? AND p.Digest = ?`)
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
			p.SetNameDigest(integrity.NameDigest{Name: name.String, Digest: digest.String})
			if !yield(p, nil) {
				break
			}
		}
		if err := rows.Err(); err != nil {
			yield(nil, err)
		}
	}
}
