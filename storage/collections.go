package storage

import (
	"context"
	"database/sql"
	"fmt"
	"iter"
	"net/url"
	"regexp"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/status"
	"github.com/wenooij/nuggit/table"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

type CollectionStore struct{ db *sql.DB }

func NewCollectionStore(db *sql.DB) *CollectionStore {
	return &CollectionStore{db: db}
}

func (s *CollectionStore) Delete(ctx context.Context, name api.NameDigest) error {
	return deleteSpec(ctx, s.db, "Collections", name)
}

func (s *CollectionStore) LoadBatch(ctx context.Context, names []api.NameDigest) iter.Seq2[*api.Collection, error] {
	return scanSpecsBatch(ctx, s.db, "Collections", names, func() *api.Collection { return new(api.Collection) })
}

func (s *CollectionStore) DeleteBatch(ctx context.Context, names []api.NameDigest) error {
	return deleteBatch(ctx, s.db, "Collections", names)
}

func (s *CollectionStore) Load(ctx context.Context, name api.NameDigest) (*api.Collection, error) {
	c := new(api.Collection)
	if err := loadSpec(ctx, s.db, "Collections", name, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CollectionStore) Store(ctx context.Context, object *api.Collection) (api.NameDigest, error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return api.NameDigest{}, err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return api.NameDigest{}, err
	}
	defer tx.Rollback()

	spec, err := marshalNullableJSONString(object)
	if err != nil {
		return api.NameDigest{}, err
	}

	prep, err := tx.PrepareContext(ctx, "INSERT INTO Collections (Name, Digest, AlwaysTrigger, Hostname, URLPattern, Spec) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return api.NameDigest{}, err
	}
	defer prep.Close()

	nameDigest, err := api.NewNameDigest(object)
	if err != nil {
		return api.NameDigest{}, err
	}

	if _, err := prep.ExecContext(ctx,
		nameDigest.GetName(),
		nameDigest.GetDigest(),
		object.GetConditions().GetAlwaysTrigger(),
		object.GetConditions().GetHostname(),
		object.GetConditions().GetURLPattern(),
		spec,
	); err != nil {
		if err, ok := err.(*sqlite.Error); ok {
			if err.Code() == sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY {
				return api.NameDigest{}, status.ErrAlreadyExists
			}
		}
		return api.NameDigest{}, err
	}

	prepPipes, err := tx.PrepareContext(ctx, fmt.Sprintf("INSERT INTO CollectionPipes (CollectionName, CollectionDigest, PipeName, PipeDigest) VALUES (%s) ON CONFLICT DO NOTHING", placeholders(4)))
	if err != nil {
		return api.NameDigest{}, err
	}
	defer prepPipes.Close()

	for _, p := range object.GetPipes() {
		if !p.HasDigest() {
			return api.NameDigest{}, fmt.Errorf("pipe referenced by collection must have a digest (%q): %w", p, status.ErrInvalidArgument)
		}
		if _, err := prepPipes.ExecContext(ctx,
			nameDigest.GetName(),
			nameDigest.GetDigest(),
			p.GetName(),
			p.GetDigest(),
		); err != nil {
			return api.NameDigest{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return api.NameDigest{}, err
	}

	return nameDigest, nil
}

func (s *CollectionStore) CreateTable(ctx context.Context, object *api.Collection, pipes []*api.Pipe) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	var tb table.Builder
	tb.Reset(object)
	if err := tb.Add(pipes...); err != nil {
		return err
	}
	createTableExpr, err := tb.Build()
	if err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, createTableExpr); err != nil {
		return err
	}

	return nil
}

func (s *CollectionStore) StoreBatch(ctx context.Context, objects []*api.Collection) ([]api.NameDigest, error) {
	// Real batch storage is not implemented yet.
	var names []api.NameDigest
	for _, o := range objects {
		name, err := s.Store(ctx, o)
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

func (s *CollectionStore) ScanNames(ctx context.Context) iter.Seq2[api.NameDigest, error] {
	return scanNames(ctx, s.db, "Collections")
}

func (s *CollectionStore) Scan(ctx context.Context) iter.Seq2[struct {
	api.NameDigest
	Elem *api.Collection
}, error] {
	return scanSpecs(ctx, s.db, "Collections", func() *api.Collection { return new(api.Collection) })
}

const triggerQuery = `SELECT
	c.Name AS CollectionName,
	c.Digest AS CollectionDigest,
    c.Spec,
	p.Name AS PipeName,
	p.Digest AS PipeDigest,
    p.Spec AS PipeSpec
FROM Collections AS c
LEFT JOIN CollectionPipes AS cp ON c.Name = cp.CollectionName AND c.Digest = cp.CollectionDigest
LEFT JOIN Pipes AS p ON p.Name = cp.PipeName AND p.Digest = cp.PipeDigest
WHERE c.Hostname = ?
    OR c.AlwaysTrigger`

func (s *CollectionStore) ScanTriggered(ctx context.Context, u *url.URL) iter.Seq2[struct {
	*api.Collection
	*api.Pipe
}, error] {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return seq2Error[struct {
			*api.Collection
			*api.Pipe
		}](err)
	}

	prep, err := conn.PrepareContext(ctx, triggerQuery)
	if err != nil {
		return seq2Error[struct {
			*api.Collection
			*api.Pipe
		}](err)
	}

	hostname := u.Hostname()
	rows, err := prep.QueryContext(ctx, hostname)
	if err != nil {
		return seq2Error[struct {
			*api.Collection
			*api.Pipe
		}](err)
	}

	urlStr := u.String()

	return func(yield func(struct {
		*api.Collection
		*api.Pipe
	}, error) bool) {
		defer conn.Close()
		defer prep.Close()
		defer rows.Close()

		zero := struct {
			*api.Collection
			*api.Pipe
		}{}
		for rows.Next() {
			var name, digest, spec, pipeName, pipeDigest, pipeSpec sql.NullString
			if err := rows.Scan(&name, &digest, &spec, &pipeName, &pipeDigest, &pipeSpec); err != nil {
				yield(zero, err)
				return
			}
			c := new(api.Collection)
			if err := unmarshalNullableJSONString(spec, c); err != nil {
				yield(zero, err)
				return
			}
			c.NameDigest = api.NameDigest{Name: name.String, Digest: digest.String}
			if c.GetConditions() == nil {
				continue // No trigger.
			}
			trigger := c.GetConditions().GetAlwaysTrigger() || c.GetConditions().GetHostname() == hostname
			if pattern := c.GetConditions().GetURLPattern(); !trigger && pattern != "" {
				match, err := regexp.MatchString(pattern, urlStr)
				if err != nil {
					yield(zero, err)
					return
				}
				trigger = match
			}
			if !trigger {
				continue // No trigger.
			}
			pipe := new(api.Pipe)
			if err := unmarshalNullableJSONString(pipeSpec, pipe); err != nil {
				yield(zero, err)
				return
			}
			pipe.NameDigest = api.NameDigest{Name: pipeName.String, Digest: pipeDigest.String}

			if !yield(struct {
				*api.Collection
				*api.Pipe
			}{c, pipe}, nil) {
				break
			}
		}
		if err := rows.Err(); err != nil {
			yield(zero, err)
		}
	}
}
