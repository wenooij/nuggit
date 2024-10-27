package storage

import (
	"context"
	"database/sql"

	"github.com/wenooij/nuggit/trigger"
)

type PlanStore struct {
	db *sql.DB
}

func NewPlanStore(db *sql.DB) *PlanStore {
	return &PlanStore{db: db}
}

func (s *PlanStore) Store(ctx context.Context, uuid string, plan *trigger.Plan) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	spec, err := marshalNullableJSONString(plan)
	if err != nil {
		return err
	}

	planResult, err := conn.ExecContext(ctx, "INSERT INTO TriggerPlans (UUID, Plan) VALUES (?, ?)", uuid, spec)
	if err != nil {
		return err
	}
	planID, err := planResult.LastInsertId()
	if err != nil {
		return err
	}

	prep, err := conn.PrepareContext(ctx, "INSERT INTO TriggerPlanPipes (PlanID, PipeID) SELECT ?, p.ID FROM Pipes AS p WHERE p.Name = ? AND p.Digest = ? LIMIT 1")
	if err != nil {
		return err
	}
	defer prep.Close()

	for _, i := range plan.GetExchanges() {
		exchange := plan.Steps[i]
		if _, err := prep.ExecContext(ctx,
			planID,
			exchange.GetOrDefaultArg("name"),
			exchange.GetOrDefaultArg("digest"),
		); err != nil {
			return err
		}
	}

	return nil
}

func (s *PlanStore) Finish(ctx context.Context, uuid string) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "UPDATE TriggerPlans SET Finished = true WHERE UUID = ?", uuid); err != nil {
		return err
	}

	return nil
}
