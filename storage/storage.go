package storage

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"log"
	"reflect"
	"strings"

	"github.com/wenooij/nuggit/integrity"
	"github.com/wenooij/nuggit/status"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

type Type = string

const (
	StorageUndefined Type = "" // Same as in memory.
	StorageInMemory  Type = "inmemory"
)

type OpStatus = string

const (
	OpUndefined OpStatus = "" // Same as StatusUnknown.
	OpUnknown   OpStatus = "unknown"
	OpStoring   OpStatus = "storing"
	OpComplete  OpStatus = "complete"
)

type UUID interface{ UUID() string }

var ErrStopScan = errors.New("stop scan")

//go:embed schema.sql
var schema string

func InitDB(ctx context.Context, db *sql.DB) error {
	log.Printf("Initializing DB...\n======== BEGIN SCHEMA ========\n%s\n======== END SCHEMA ========\n", schema)
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, schema); err != nil {
		return err
	}
	return nil
}

func marshalNullableJSONString(x any) (sql.NullString, error) {
	if reflect.ValueOf(x).IsZero() {
		return sql.NullString{}, nil
	}
	var sb strings.Builder
	e := json.NewEncoder(&sb)
	if err := e.Encode(x); err != nil {
		return sql.NullString{}, err
	}
	s := sb.String()
	return sql.NullString{String: s, Valid: true}, nil
}

func unmarshalNullableJSONString(data sql.NullString, x any) error {
	if !data.Valid {
		return nil
	}
	return json.Unmarshal([]byte(data.String), x)
}

// n placeholders: ?,?,?,...
func placeholders(n int) string {
	var sb strings.Builder
	sb.Grow(2 * n)
	for ; n > 1; n-- {
		sb.WriteString("?,")
	}
	if n == 1 {
		sb.WriteByte('?')
	}
	return sb.String()
}

func convertToAnySlice[E interface {
	sql.Null[bool]
	sql.Null[string]
}](es []E) []any {
	res := make([]any, len(es))
	for i, e := range es {
		res[i] = e
	}
	return res
}

func convertNamesToAnySlice(es []integrity.NameDigest) []any {
	res := make([]any, len(es))
	for i, e := range es {
		nameDigest, err := integrity.FormatString(e)
		if err != nil {
			panic(err)
		}
		res[i] = nameDigest
	}
	return res
}

var validTableNames = map[string]struct{}{
	"Views":        {},
	"ViewPipes":    {},
	"ViewsData":    {},
	"Pipes":        {},
	"PipeVersions": {},
	"Triggers":     {},
}

func safeTableName(s string) string {
	if _, ok := validTableNames[s]; ok {
		return s
	}
	panic(fmt.Sprintf("Unexpected table name (%q)", s))
}

func loadSpec[E integrity.CheckDigestable](ctx context.Context, db *sql.DB, tableName string, nameDigest integrity.NameDigest, e E) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	prep, err := conn.PrepareContext(ctx, fmt.Sprintf("SELECT t.Name, t.Digest, t.Spec FROM %s AS t WHERE t.Name = ? AND t.Digest = ? LIMIT 1", safeTableName(tableName)))
	if err != nil {
		return err
	}
	defer prep.Close()

	var name, digest, spec sql.NullString
	if err := prep.QueryRowContext(ctx, nameDigest.GetName(), nameDigest.GetDigest()).Scan(&name, &digest, &spec); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return status.ErrNotFound
		}
		return err
	}
	if err := unmarshalNullableJSONString(spec, e); err != nil {
		return err
	}

	integrity.SetName(e, name.String)
	if err := integrity.SetCheckDigest(e, digest.String); err != nil {
		return fmt.Errorf("failed to set digest (%q): %w", name.String, err)
	}
	return nil
}

func deleteSpec(ctx context.Context, db *sql.DB, tableName string, nameDigest integrity.NameDigest) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	prep, err := conn.PrepareContext(ctx, fmt.Sprintf("DELETE FROM %s AS t WHERE t.Name = ? AND t.Digest = ?", safeTableName(tableName)))
	if err != nil {
		return err
	}
	defer prep.Close()

	if _, err := prep.ExecContext(ctx, nameDigest.GetName(), nameDigest.GetDigest()); err != nil {
		return err
	}
	return nil
}

func deleteBatch(ctx context.Context, db *sql.DB, tableName string, names []integrity.NameDigest) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	prep, err := tx.PrepareContext(ctx, fmt.Sprintf("DELETE FROM %s AS t WHERE CONCAT (t.Name, '@', t.Digest) IN (%s)", safeTableName(tableName), placeholders(len(names))))
	if err != nil {
		return err
	}
	defer prep.Close()

	if _, err := prep.ExecContext(ctx, convertNamesToAnySlice(names)...); err != nil {
		return err
	}
	return tx.Commit()
}

func seq2Error[E any](err error) iter.Seq2[E, error] {
	return func(yield func(E, error) bool) {
		var zero E
		yield(zero, err)
	}
}

func scanSpecsBatch[E integrity.CheckDigestable](ctx context.Context, db *sql.DB, tableName string, names []integrity.NameDigest, newE func() E) iter.Seq2[E, error] {
	conn, err := db.Conn(ctx)
	if err != nil {
		return seq2Error[E](err)
	}

	prep, err := conn.PrepareContext(ctx, fmt.Sprintf(`SELECT
	t.Name,
	t.Digest,
	t.Spec
FROM %s AS t
WHERE CONCAT(t.Name, '@', t.Digest) IN (%s)`, safeTableName(tableName), placeholders(len(names))))
	if err != nil {
		return seq2Error[E](err)
	}

	rows, err := prep.QueryContext(ctx, convertNamesToAnySlice(names)...)
	if err != nil {
		return seq2Error[E](err)
	}

	return func(yield func(E, error) bool) {
		defer conn.Close()
		defer prep.Close()
		defer rows.Close()

		var zero E
		for rows.Next() {
			var name, digest, spec sql.NullString
			if err := rows.Scan(&name, &digest, &spec); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					continue // Name is missing, skip it.
				}
				yield(zero, err)
				return
			}
			e := newE()
			if err := unmarshalNullableJSONString(spec, e); err != nil {
				yield(zero, err)
				return
			}
			integrity.SetName(e, name.String)
			if err := integrity.SetCheckDigest(e, digest.String); err != nil {
				yield(zero, fmt.Errorf("failed to set digest (%q): %w", name.String, err))
				return
			}
			if !yield(e, nil) {
				break
			}
		}
		if err := rows.Err(); err != nil {
			yield(zero, err)
		}
	}
}

func scanNames(ctx context.Context, db *sql.DB, tableName string) iter.Seq2[integrity.NameDigest, error] {
	conn, err := db.Conn(ctx)
	if err != nil {
		return seq2Error[integrity.NameDigest](err)
	}

	rows, err := conn.QueryContext(ctx, fmt.Sprintf("SELECT t.Name, t.Digest FROM %s AS t", safeTableName(tableName)))
	if err != nil {
		return seq2Error[integrity.NameDigest](err)
	}

	return func(yield func(integrity.NameDigest, error) bool) {
		defer conn.Close()
		defer rows.Close()

		for rows.Next() {
			var name, digest sql.NullString
			if err := rows.Scan(&name, &digest); err != nil {
				yield(nil, err)
				return
			}
			if !yield(integrity.KeyLit(name.String, digest.String), nil) {
				break
			}
		}
		if err := rows.Err(); err != nil {
			yield(nil, err)
		}
	}
}

func scanSpecs[E integrity.CheckDigestable](ctx context.Context, db *sql.DB, tableName string, newE func() E) iter.Seq2[E, error] {
	conn, err := db.Conn(ctx)
	if err != nil {
		return seq2Error[E](err)
	}

	rows, err := conn.QueryContext(ctx, fmt.Sprintf("SELECT t.Name, t.Digest, t.Spec FROM %s AS t", safeTableName(tableName)))
	if err != nil {
		return seq2Error[E](err)
	}

	return func(yield func(E, error) bool) {
		defer conn.Close()
		defer rows.Close()

		var zero E
		for rows.Next() {
			var name, digest, spec sql.NullString
			if err := rows.Scan(&name, &digest, &spec); err != nil {
				yield(zero, err)
				return
			}
			e := newE()
			if err := unmarshalNullableJSONString(spec, e); err != nil {
				yield(zero, err)
				return
			}
			e.SetName(name.String)
			if err := integrity.SetCheckDigest(e, digest.String); err != nil {
				yield(zero, fmt.Errorf("failed to set digest (%q): %w", name.String, err))
				return
			}
			if !yield(e, nil) {
				break
			}
		}
		if err := rows.Err(); err != nil {
			yield(zero, err)
		}
	}
}

func handleAlreadyExists(object string, key integrity.NameDigest, err error) error {
	if err, ok := err.(*sqlite.Error); ok {
		if err.Code() == sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY || err.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
			return fmt.Errorf("failed to store %s (%q): %w", object, integrity.Key(key), status.ErrAlreadyExists)
		}
	}
	return err
}

func ignoreAlreadyExists(err error) error {
	if err, ok := err.(*sqlite.Error); ok {
		if err.Code() == sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY || err.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
			return nil
		}
	}
	return err
}
