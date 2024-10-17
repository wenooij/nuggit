package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/status"
)

type CollectionStore struct{ db *sql.DB }

func NewCollectionStore(db *sql.DB) *CollectionStore {
	return &CollectionStore{db: db}
}

func (s *CollectionStore) Len(ctx context.Context) (int, bool) {
	rows, err := s.db.QueryContext(ctx, "SELECT COUNT(*) FROM Collections")
	if err != nil {
		log.Printf("Failed to query CollectionStore.Len: %v", err)
		return 0, false
	}
	return lenRows(rows)
}

func (s *CollectionStore) Delete(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM Collections WHERE CollectionID = ?", id); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM CollectionConditions WHERE CollectionID = ?", id); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	return status.ErrUnimplemented
}

func (s *CollectionStore) Load(ctx context.Context, id string) (*api.CollectionRich, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	row := tx.QueryRowContext(ctx, "SELECT * FROM Collections WHERE CollectionID = ?", id)

	cb := &api.CollectionBase{}
	var condID int64
	var numPoints int

	if err := row.Scan(&cb.Name, &cb.DryRun, &cb.IncludeMetadata, &condID, &numPoints); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.ErrNotFound
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &api.CollectionRich{
		Collection: &api.Collection{
			CollectionLite: api.NewCollectionLite(id),
			CollectionBase: cb,
		},
	}, nil
}

func (s *CollectionStore) Store(ctx context.Context, object *api.CollectionRich) (err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	id := object.UUID()
	row := tx.QueryRowContext(ctx, "SELECT 1 FROM Collections WHERE CollectionID = ?", id)
	var i int64
	if err := row.Scan(&i); err == nil {
		_ = tx.Rollback()
		return status.ErrAlreadyExists
	} else if errors.Is(err, sql.ErrNoRows) {
		log.Println("calling storeOrReplaceCollectionTx")
		return s.storeOrReplaceCollectionTx(ctx, tx, object)
	} else {
		_ = tx.Rollback()
		return err
	}
}

func (s *CollectionStore) StoreOrReplace(ctx context.Context, object *api.CollectionRich) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	return s.storeOrReplaceCollectionTx(ctx, tx, object)
}

func (s *CollectionStore) storeOrReplaceCollectionTx(ctx context.Context, tx *sql.Tx, object *api.CollectionRich) error {
	id := object.UUID()

	var conditions *string
	if cond := object.Conditions; cond != nil {
		data, err := json.Marshal(cond)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		conditions = new(string)
		*conditions = string(data)
	}

	var ss *string
	if state := object.State; state != nil {
		data, err := json.Marshal(object.State)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		ss = new(string)
		*ss = string(data)
	}

	var points *string
	if pts := object.Points; len(pts) > 0 {
		data, err := json.Marshal(object.Points)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		points = new(string)
		*points = string(data)
	}

	if _, err := tx.ExecContext(ctx, "INSERT INTO Collections (CollectionID, Name, DryRun, IncludeMetadata, Conditions, State, Points) VALUES (?, ?, ?, ?, ?, ?, ?)",
		id,
		object.Name,
		object.DryRun,
		object.IncludeMetadata,
		conditions,
		ss,
		points); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *CollectionStore) Scan(ctx context.Context, scanFn func(*api.CollectionLite, error) error) error {
	rows, err := s.db.QueryContext(ctx, "SELECT CollectionID FROM Collections")
	if err != nil {
		return err
	}
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err := scanFn(api.NewCollectionLite(id), err); err != nil {
			if err == ErrStopScan {
				return nil
			}
			return err
		}
	}
	return nil
}
