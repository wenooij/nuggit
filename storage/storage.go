package storage

import (
	"database/sql"
	_ "embed"
	"errors"
	"log"
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

func InitDB(db *sql.DB) error {
	log.Printf("Initializing DB...\n======== BEGIN SCHEMA ========\n%s\n======== END SCHEMA ========\n", schema)
	if _, err := db.Exec(schema); err != nil {
		return err
	}
	return nil
}

func lenRows(rows *sql.Rows) (int, bool) {
	count := 0
	for ; rows.Next(); count++ {
	}
	if err := rows.Err(); err != nil {
		return count, false
	}
	return count, true
}
