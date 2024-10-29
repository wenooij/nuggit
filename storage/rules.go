package storage

import (
	"context"
	"database/sql"
	"fmt"
	"iter"
	"net/url"
	"regexp"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/integrity"
)

type RuleStore struct {
	db *sql.DB
}

func NewRuleStore(db *sql.DB) *RuleStore {
	return &RuleStore{db}
}

func (s *RuleStore) StoreRule(ctx context.Context, rule nuggit.Rule) error {
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

	if _, err := tx.ExecContext(ctx, "INSERT OR IGNORE INTO Rules (Hostname, URLPattern, AlwaysTrigger, Disable) VALUES (?, ?, ?, ?) ON CONFLICT DO NOTHING",
		rule.Hostname,
		rule.URLPattern,
		rule.AlwaysTrigger,
		rule.Disable); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM RuleLabels WHERE RuleID IN (
	SELECT r.ID
	FROM Rules AS r
	WHERE r.Hostname = ? AND r.URLPattern = ? AND r.AlwaysTrigger = ? AND Disable = ?
)`, rule.Hostname, rule.URLPattern, rule.AlwaysTrigger, rule.Disable); err != nil {
		return err
	}

	prep, err := conn.PrepareContext(ctx, `INSERT INTO RuleLabels (RuleID, Label)
SELECT r.ID, ?
FROM Rules AS r
WHERE r.Hostname = ? AND r.URLPattern = ? AND r.AlwaysTrigger = ? AND Disable = ?
LIMIT 1`)
	if err != nil {
		return err
	}

	for _, label := range rule.Labels {
		if _, err := prep.ExecContext(ctx, label, rule.Hostname, rule.URLPattern, rule.AlwaysTrigger, rule.Disable); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *RuleStore) DeleteRule(ctx context.Context, rule nuggit.Rule) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `DELETE FROM Rules AS r
WHERE r.Hostname = ? AND r.URLPattern = ? AND r.AlwaysTrigger = ?
LIMIT 1`,
		rule.Hostname,
		rule.URLPattern,
		rule.AlwaysTrigger); err != nil {
		return err
	}

	return nil
}

const triggerQuery = `SELECT 
    p.Name,
    p.Digest,
    p.Spec,
    MAX(u.URLPattern) AS URLPattern,
    COALESCE(MAX(u.AlwaysTrigger), FALSE) AS AlwaysTrigger
FROM Pipes AS p
JOIN Resources AS r ON p.ID = r.PipeID
JOIN ResourceLabels AS rl ON r.ID = rl.ResourceID
JOIN RuleLabels AS ul ON rl.Label = ul.Label
JOIN Rules AS u ON ul.RuleID = u.ID
WHERE NOT r.ID IN (
    SELECT r.ID
    FROM Resources AS r
    JOIN ResourceLabels AS rl ON r.ID = rl.ResourceID
    WHERE rl.Label = 'disabled'
)
GROUP BY p.Name, p.Digest, p.Spec
HAVING NOT COALESCE(MAX(u.Disable), FALSE) AND (MAX(u.AlwaysTrigger) OR MAX(u.Hostname) = ?)`

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

		for rows.Next() {
			var name, digest, spec, urlPattern sql.NullString
			var alwaysTrigger sql.NullBool
			if err := rows.Scan(&name, &digest, &spec, &urlPattern, &alwaysTrigger); err != nil {
				yield(nil, err)
				return
			}

			pipeSpec := new(nuggit.Pipe)
			if err := unmarshalNullableJSONString(spec, pipeSpec); err != nil {
				yield(nil, err)
				return
			}
			pipe := new(api.Pipe)
			pipe.Pipe = *pipeSpec

			if err := integrity.SetCheckNameDigest(pipe, name.String, digest.String); err != nil {
				yield(nil, fmt.Errorf("failed to set digest (%q): %w", name.String, err))
				return
			}

			if !alwaysTrigger.Bool && urlPattern.String != "" {
				// Test URL pattern since its not empty.
				match, err := regexp.MatchString(urlPattern.String, urlStr)
				if err != nil {
					yield(nil, err)
					return
				}
				if !match {
					continue // This rule failed to match the URL.
				}
			}

			// Pipe has been triggered either by AlwaysTrigger
			// Or a matching Hostname and URL pattern.
			if !yield(pipe, nil) {
				break
			}
		}
		if err := rows.Err(); err != nil {
			yield(nil, err)
		}
	}
}
