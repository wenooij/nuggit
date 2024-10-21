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

	"github.com/google/uuid"
	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/status"
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

func newUUID() (string, error) {
	u, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("%v: %w", err, status.ErrInternal)
	}
	id := u.String()
	return id, nil
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

func convertNamesToAnySlice(es []api.NameDigest) []any {
	res := make([]any, len(es))
	for i, e := range es {
		res[i] = e.String()
	}
	return res
}

var validTableNames = map[string]struct{}{
	"Collections":     {},
	"CollectionPipes": {},
	"CollectionsData": {},
	"Pipes":           {},
	"PipeVersions":    {},
	"Triggers":        {},
	"TriggerResults":  {},
}

func safeTableName(s string) string {
	if _, ok := validTableNames[s]; ok {
		return s
	}
	panic(fmt.Sprintf("Unexpected table name (%q)", s))
}

func loadSpec[T interface{ GetName() string }](ctx context.Context, db *sql.DB, tableName string, nameDigest api.NameDigest, instance T) error {
	tableName = safeTableName(tableName)

	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	prep, err := conn.PrepareContext(ctx, fmt.Sprintf("SELECT t.Spec FROM %s AS t WHERE t.Name = ? AND t.Digest = ? LIMIT 1", tableName))
	if err != nil {
		return err
	}
	defer prep.Close()

	spec := sql.NullString{}
	if err := prep.QueryRowContext(ctx, nameDigest.GetName(), nameDigest.GetDigest()).Scan(&spec); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return status.ErrNotFound
		}
		return err
	}
	if err := unmarshalNullableJSONString(spec, instance); err != nil {
		return err
	}
	// TODO: Set NameDigest for the new object.
	return nil
}

func deleteSpec(ctx context.Context, db *sql.DB, tableName string, nameDigest api.NameDigest) error {
	tableName = safeTableName(tableName)

	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	prep, err := conn.PrepareContext(ctx, fmt.Sprintf("DELETE FROM %s AS t WHERE t.Name = ? AND t.Digest = ?", tableName))
	if err != nil {
		return err
	}
	defer prep.Close()

	if _, err := prep.ExecContext(ctx, nameDigest.GetName(), nameDigest.GetDigest()); err != nil {
		return err
	}
	return nil
}

func deleteBatch(ctx context.Context, db *sql.DB, tableName string, names []api.NameDigest) error {
	tableName = safeTableName(tableName)

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

	prep, err := tx.PrepareContext(ctx, fmt.Sprintf("DELETE FROM %s AS t WHERE CONCAT (t.Name, '@', t.Digest) IN (%s)", tableName, placeholders(len(names))))
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

func scanSpecsBatch[T interface{ GetName() string }](ctx context.Context, db *sql.DB, tableName string, names []api.NameDigest, newT func() T) iter.Seq2[T, error] {
	tableName = safeTableName(tableName)

	conn, err := db.Conn(ctx)
	if err != nil {
		return seq2Error[T](err)
	}
	defer conn.Close()

	prep, err := conn.PrepareContext(ctx, fmt.Sprintf(`SELECT
	p.Spec
FROM %s AS t
WHERE CONCAT(t.Name, '@', t.Digest) IN (%s)`, tableName, placeholders(len(names))))
	if err != nil {
		return seq2Error[T](err)
	}

	rows, err := prep.QueryContext(ctx, convertNamesToAnySlice(names)...)
	if err != nil {
		return seq2Error[T](err)
	}

	return func(yield func(T, error) bool) {
		defer prep.Close()
		defer rows.Close()

		var zero T
		for rows.Next() {
			spec := sql.NullString{}
			if err := rows.Scan(&spec); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					continue // Name is missing, skip it.
				}
				yield(zero, err)
				return
			}
			t := newT()
			if err := unmarshalNullableJSONString(spec, t); err != nil {
				yield(zero, err)
				return
			}
			if !yield(t, nil) {
				break
			}
		}
		if err := rows.Err(); err != nil {
			yield(zero, err)
		}
	}
}

func scanNames(ctx context.Context, db *sql.DB, tableName string) iter.Seq2[api.NameDigest, error] {
	tableName = safeTableName(tableName)

	conn, err := db.Conn(ctx)
	if err != nil {
		return seq2Error[api.NameDigest](err)
	}

	rows, err := conn.QueryContext(ctx, fmt.Sprintf("SELECT t.Name, t.Digest FROM %s AS t", tableName))
	if err != nil {
		return seq2Error[api.NameDigest](err)
	}

	return func(yield func(api.NameDigest, error) bool) {
		defer conn.Close()
		defer rows.Close()

		for rows.Next() {
			var name sql.NullString
			var digest sql.NullString
			if err := rows.Scan(&name, &digest); err != nil {
				yield(api.NameDigest{}, err)
				return
			}
			if !yield(api.NameDigest{Name: name.String, Digest: digest.String}, nil) {
				break
			}
		}
		if err := rows.Err(); err != nil {
			yield(api.NameDigest{}, err)
		}
	}
}

func scanSpecs[T interface{ GetName() string }](ctx context.Context, db *sql.DB, tableName string, newT func() T) iter.Seq2[struct {
	api.NameDigest
	Elem T
}, error] {
	tableName = safeTableName(tableName)

	conn, err := db.Conn(ctx)
	if err != nil {
		return seq2Error[struct {
			api.NameDigest
			Elem T
		}](err)
	}

	rows, err := conn.QueryContext(ctx, fmt.Sprintf("SELECT t.Spec FROM %s AS t", tableName))
	if err != nil {
		return seq2Error[struct {
			api.NameDigest
			Elem T
		}](err)
	}

	return func(yield func(struct {
		api.NameDigest
		Elem T
	}, error) bool) {
		defer conn.Close()
		defer rows.Close()

		var zt T
		zero := struct {
			api.NameDigest
			Elem T
		}{api.NameDigest{}, zt}
		for rows.Next() {
			var spec sql.NullString
			if err := rows.Scan(&spec); err != nil {
				yield(zero, err)
				return
			}
			t := newT()
			if err := unmarshalNullableJSONString(spec, t); err != nil {
				yield(zero, err)
				return
			}
			// TODO: Set NameDigest for the new object.
			if !yield(zero, nil) {
				break
			}
		}
		if err := rows.Err(); err != nil {
			yield(zero, err)
		}
	}
}
