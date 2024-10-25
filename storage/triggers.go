package storage

import (
	"context"
	"database/sql"
	"iter"
	"net/url"
	"regexp"

	"github.com/wenooij/nuggit/api"
)

type CriteriaStore struct {
	db *sql.DB
}

func NewCriteriaStore(db *sql.DB) *CriteriaStore {
	return &CriteriaStore{db}
}

func (s *CriteriaStore) Disable(ctx context.Context, nameDigest api.NameDigest) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `UPDATE OR IGNORE TriggerCriteria AS t SET Disabled = TRUE WHERE EXISTS (
	SELECT 1
	FROM Pipes AS p
	WHERE p.Name = ? AND p.Digest = ? AND p.CriteriaID = t.ID
	LIMIT 1
)`); err != nil {
		return err
	}

	return nil
}

func (s *CriteriaStore) Store(ctx context.Context, t *api.TriggerCriteria) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "INSERT INTO TriggerCriteria (AlwaysTrigger, Hostname, URLPattern) VALUES (?, ?, ?)",
		t.GetAlwaysTrigger(),
		t.GetHostname(),
		t.GetURLPattern()); err != nil {
		return err
	}

	return nil
}

const triggerQuery = `SELECT
	p.Name AS PipeName,
	p.Digest AS PipeDigest,
    p.Spec AS PipeSpec,
	t.AlwaysTrigger,
	t.URLPattern,
FROM Pipes AS p
LEFT JOIN TriggerCriteria AS t ON p.CriteriaID = t.ID
WHERE t.Hostname = ? OR t.AlwaysTrigger`

func (s *CriteriaStore) ScanMatched(ctx context.Context, u *url.URL) iter.Seq2[*api.Pipe, error] {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return seq2Error[*api.Pipe](err)
	}

	prep, err := conn.PrepareContext(ctx, triggerQuery)
	if err != nil {
		return seq2Error[*api.Pipe](err)
	}

	hostname := u.Hostname()
	rows, err := prep.QueryContext(ctx, hostname)
	if err != nil {
		return seq2Error[*api.Pipe](err)
	}

	urlStr := u.String()

	return func(yield func(*api.Pipe, error) bool) {
		defer conn.Close()
		defer prep.Close()
		defer rows.Close()

		for rows.Next() {
			var name, digest, spec, urlPattern sql.NullString
			var alwaysTrigger sql.NullBool
			if err := rows.Scan(&name, &digest, &spec, &alwaysTrigger, &urlPattern); err != nil {
				yield(nil, err)
				return
			}

			pipe := new(api.Pipe)
			if err := unmarshalNullableJSONString(spec, pipe); err != nil {
				yield(nil, err)
				return
			}
			pipe.SetNameDigest(api.NameDigest{Name: name.String, Digest: digest.String})

			if !alwaysTrigger.Bool && urlPattern.Valid {
				// Test URL pattern since its not null.
				match, err := regexp.MatchString(urlPattern.String, urlStr)
				if err != nil {
					yield(nil, err)
					return
				}
				if !match {
					continue
				}
			}

			// Pipe has been triggered either by always_trigger
			// Or a matching Hostname or URL pattern.
			if !yield(pipe, nil) {
				break
			}
		}
		if err := rows.Err(); err != nil {
			yield(nil, err)
		}
	}
}

type PlanStore struct {
	db *sql.DB
}

func NewPlanStore(db *sql.DB) *PlanStore {
	return &PlanStore{db: db}
}

func (s *PlanStore) Store(ctx context.Context, uuid string, plan *api.TriggerPlan) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	spec, err := marshalNullableJSONString(plan)
	if err != nil {
		return err
	}

	if _, err := conn.ExecContext(ctx, "INSERT INTO TriggerPlans (UUID, Plan) VALUES (?, ?)",
		uuid, spec); err != nil {
		return err
	}

	return nil
}

func (s *PlanStore) Finish(ctx context.Context, uuid string) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "UPDATE TriggerPlans SET Finished = true WHERE UUID = ?", uuid); err != nil {
		return err
	}

	return nil
}
