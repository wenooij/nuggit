package storage

import (
	"context"
	"database/sql"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/integrity"
	"github.com/wenooij/nuggit/points"
)

type ResultStore struct {
	db *sql.DB
}

func NewResultStore(db *sql.DB) *ResultStore {
	return &ResultStore{db: db}
}

type nextStop struct {
	next func() (any, error, bool)
	stop func()
}

func (s *ResultStore) StoreResults(ctx context.Context, event *api.TriggerEvent, results []api.TriggerResult) error {
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

	var planID int64
	if err := tx.QueryRowContext(ctx, "SELECT ID FROM Plans WHERE UUID = ? LIMIT 1", event.Plan).Scan(&planID); err != nil {
		return err
	}

	triggerResult, err := tx.ExecContext(ctx, `INSERT INTO Events (PlanID, Implicit, URL, Timestamp) VALUES (?, ?, ?, ?)`,
		planID,
		event.GetImplicit(),
		event.GetURL(),
		event.GetTimestamp())
	if err != nil {
		return err // No unique columns or AlreadyExists to handle here.
	}
	eventID, err := triggerResult.LastInsertId()
	if err != nil {
		return err
	}

	prep, err := tx.PrepareContext(ctx, `INSERT INTO Results (EventID, PipeID, SequenceID, TypeNumber, Result)
SELECT ?, p.ID, ?, p.TypeNumber, ?
FROM Pipes AS p WHERE p.Name = ? AND p.Digest = ?
LIMIT 1`)
	if err != nil {
		return err // Let AlreadyExists fail as this would indicate an issue with the sequencing logic.
	}
	defer prep.Close()

	for _, res := range results {
		var p nuggit.Point
		p.Scalar = res.Scalar
		var seq int
		for v, err := range points.Values(p, res.Result) {
			if err != nil {
				return err
			}
			nameDigest, err := integrity.ParseNameDigest(res.Pipe)
			if err != nil {
				return err
			}
			if _, err := prep.ExecContext(ctx,
				eventID,
				seq,
				v,
				nameDigest.GetName(),
				nameDigest.GetDigest(),
			); err != nil {
				return err
			}
			seq++
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
