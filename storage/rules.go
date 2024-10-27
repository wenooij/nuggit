package storage

import (
	"context"
	"database/sql"
	"iter"
	"net/url"
	"regexp"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/integrity"
	"github.com/wenooij/nuggit/trigger"
)

type RuleStore struct {
	db *sql.DB
}

func NewRuleStore(db *sql.DB) *RuleStore {
	return &RuleStore{db}
}

func (s *RuleStore) Disable(ctx context.Context, nameDigest integrity.NameDigest) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `UPDATE OR IGNORE TriggerRule AS t SET Disabled = TRUE WHERE EXISTS (
	SELECT 1
	FROM Pipes AS p
	WHERE p.Name = ? AND p.Digest = ? AND p.RuleID = t.ID
	LIMIT 1
)`); err != nil {
		return err
	}

	return nil
}

func (s *RuleStore) StoreRule(ctx context.Context, pipe integrity.NameDigest, rule *trigger.Rule) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "INSERT INTO TriggerRules (Hostname, URLPattern) VALUES (?, ?) ON CONFLICT DO NOTHING",
		rule.GetHostname(),
		rule.GetURLPattern()); err != nil {
		return err
	}

	if _, err := conn.ExecContext(ctx, `INSERT INTO PipeRules (PipeID, RuleID)
SELECT p.ID, r.ID
FROM Pipes AS p
JOIN TriggerRules AS r ON
	r.Hostname = ? AND COALESCE(r.URLPattern, '') = ?
WHERE p.Name = ? AND p.Digest = ?
LIMIT 1`,
		rule.GetHostname(),
		rule.GetURLPattern(),
		pipe.GetName(),
		pipe.GetDigest()); err != nil {
		return err
	}

	return nil
}

const triggerQuery = `SELECT
    p.Name AS PipeName,
    p.Digest AS PipeDigest,
    p.Spec AS PipeSpec,
    p.AlwaysTrigger,
    tr.URLPattern
FROM Pipes AS p
LEFT JOIN PipeRules AS pr ON pr.PipeID = p.ID
LEFT JOIN TriggerRules AS tr ON pr.RuleID = tr.ID
WHERE NOT p.Disabled AND (p.AlwaysTrigger OR tr.Hostname = ?)`

func (s *RuleStore) ScanMatched(ctx context.Context, u *url.URL) iter.Seq2[*api.Pipe, error] {
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

		distinct := map[integrity.NameDigest]struct{}{}

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

			nameDigest := integrity.NameDigest{Name: name.String, Digest: digest.String}
			pipe.SetNameDigest(nameDigest)

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

			// In case multiple rules have matched,
			// Ensure we yield each pipe no more than once.
			// TODO: Think of structural ways to avoid this.
			if _, found := distinct[nameDigest]; found {
				continue
			}
			distinct[nameDigest] = struct{}{}

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
