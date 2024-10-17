package storage

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"strings"
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

func convertToAnySlice[E any](es []E) []any {
	res := make([]any, len(es))
	for i, e := range es {
		res[i] = e
	}
	return res
}
