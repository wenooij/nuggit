package storage

import (
	"context"
	"database/sql"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/integrity"
	"github.com/wenooij/nuggit/table"
)

type ViewStore struct{ db *sql.DB }

func NewViewStore(db *sql.DB) *ViewStore {
	return &ViewStore{db: db}
}

func (s *ViewStore) Store(ctx context.Context, uuid string, view *api.View) error {
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

	spec, err := marshalNullableJSONString(view.GetSpec())
	if err != nil {
		return err
	}

	viewResult, err := tx.ExecContext(ctx, "INSERT INTO Views (UUID, Spec) VALUES (?, ?)",
		uuid, spec)
	if err != nil {
		return err // We do not expect AlwaysExists here.
	}
	viewID, err := viewResult.LastInsertId()
	if err != nil {
		return err
	}

	prep, err := conn.PrepareContext(ctx, `INSERT OR IGNORE INTO ViewPipes (ViewID, PipeID)
SELECT ?, p.ID
FROM Pipes AS p
WHERE p.Name = ? AND p.Digest = ? LIMIT 1`)
	if err != nil {
		return err
	}
	defer prep.Close()

	for _, p := range view.GetColumns() {
		pipe, err := integrity.ParseNameDigest(p.Pipe)
		if err != nil {
			return err
		}
		if _, err := prep.ExecContext(ctx, viewID, pipe.GetName(), pipe.GetDigest()); err != nil {
			return err
		}
	}

	if err := s.createViewTx(ctx, tx, uuid, view); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *ViewStore) createViewTx(ctx context.Context, tx *sql.Tx, uuid string, view *api.View) error {
	var vb table.ViewBuilder
	vb.Reset()
	if err := vb.SetView(uuid, view.GetAlias()); err != nil {
		return err
	}
	for _, col := range view.GetColumns() {
		vb.AddViewColumn(col)
	}
	createViewsExpr, err := vb.Build()
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, createViewsExpr); err != nil {
		return err
	}
	return nil
}
