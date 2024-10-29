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

	if _, err := tx.ExecContext(ctx, "INSERT INTO Rules (Hostname, URLPattern) VALUES (?, ?) ON CONFLICT DO NOTHING",
		rule.GetHostname(),
		rule.GetURLPattern()); err != nil {
		return ignoreAlreadyExists(err) // Currently trigger rules are fully unique tables.
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM RuleLabels WHERE RuleID IN (
	SELECT r.ID
	FROM Rules AS r
	WHERE r.Hostname = ? AND r.URLPattern = ?
)`, rule.Hostname, rule.URLPattern); err != nil {
		return err
	}

	prep, err := conn.PrepareContext(ctx, `INSERT INTO RuleLabels (RuleID, Label)
SELECT r.ID, ?
FROM Rules AS r
WHERE r.Hostname = ? AND r.URLPattern = ?
LIMIT 1`)
	if err != nil {
		return err
	}

	for _, label := range rule.Labels {
		if _, err := prep.ExecContext(ctx, label, rule.Hostname, rule.URLPattern); err != nil {
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
WHERE r.Hostname = ? AND r.URLPattern = ?
LIMIT 1`,
		rule.GetHostname(),
		rule.GetURLPattern()); err != nil {
		return err
	}

	return nil
}

const triggerQuery = `SELECT 
    p.Name,
    p.Digest,
    p.Spec,
    MAX(p.AlwaysTrigger) AS AlwaysTrigger,
    iif(COUNT(u.URLPattern) = COUNT(*), u.URLPattern, NULL) AS URLPattern
FROM Pipes AS p
LEFT JOIN Resources AS r ON p.ID = r.PipeID
LEFT JOIN ResourceLabels AS rl ON r.ID = rl.ResourceID
LEFT JOIN RuleLabels AS ul ON EXISTS (
    SELECT 1
    FROM RuleLabels AS ul
    WHERE rl.Label = ul.Label
)
LEFT JOIN Rules AS u ON ul.RuleID = u.ID
WHERE NOT COALESCE(p.Disabled, FALSE) AND (p.AlwaysTrigger OR u.Hostname = ?)
GROUP BY p.Name, p.Digest`

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

			integrity.SetName(pipe, name.String)
			if err := integrity.SetCheckDigest(pipe, digest.String); err != nil {
				yield(nil, fmt.Errorf("failed to set digest (%q): %w", name.String, err))
				return
			}

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
			key := integrity.Key(pipe)
			if _, found := distinct[key]; found {
				continue
			}
			distinct[key] = struct{}{}

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
