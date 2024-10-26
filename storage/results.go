package storage

import (
	"context"
	"database/sql"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/api"
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
	if err := tx.QueryRowContext(ctx, "SELECT ID FROM TriggerPlans WHERE UUID = ? LIMIT 1", event.Plan).Scan(&planID); err != nil {
		return err
	}

	triggerResult, err := tx.ExecContext(ctx, `INSERT INTO TriggerEvents (PlanID, Implicit, URL, Timestamp) VALUES (?, ?, ?, ?)`,
		planID,
		event.GetImplicit(),
		event.GetURL(),
		event.GetTimestamp())
	if err != nil {
		return err
	}
	eventID, err := triggerResult.LastInsertId()
	if err != nil {
		return err
	}

	prep, err := tx.PrepareContext(ctx, `INSERT INTO TriggerResults (EventID, PipeID, TypeNumber, Result)
SELECT ?, p.ID, p.TypeNumber, ?
FROM Pipes AS p WHERE p.Name = ? AND p.Digest = ?
LIMIT 1`)
	if err != nil {
		return err
	}
	defer prep.Close()

	for _, res := range results {
		// TODO: FIXME: We need to get point information for each result.
		// I haven't decided how best to do that yet.
		// For now assume everything is just bytes.
		var p nuggit.Point
		for v, err := range points.UnmarshalFlat(p, res.Result) {
			if err != nil {
				return err
			}
			if _, err := prep.ExecContext(ctx, eventID, v, res.Pipe.GetName(), res.Pipe.GetDigest()); err != nil {
				return err
			}
		}
	}

	return nil
}
