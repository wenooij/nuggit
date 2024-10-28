package storage

import (
	"context"
	"database/sql"

	"github.com/wenooij/nuggit/api"
)

type ResourceStore struct {
	db *sql.DB
}

func NewResourceStore(db *sql.DB) *ResourceStore {
	return &ResourceStore{db}
}

func (s *ResourceStore) StorePipeResource(ctx context.Context, resource *api.Resource, pipe *api.Pipe) error {
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

	resourceResult, err := tx.ExecContext(ctx, `INSERT INTO Resources (APIVersion, Kind, Version, Description, PipeID)
SELECT ?, ?, ?, ?, p.ID
FROM Pipes AS p
WHERE p.Name = ? AND p.Digest = ?
LIMIT 1`,
		resource.GetAPIVersion(),
		resource.GetKind(),
		resource.GetMetadata().GetVersion(),
		resource.GetMetadata().GetDescription(),
		pipe.GetName(),
		pipe.GetDigest(),
	)
	if err != nil {
		return handleAlreadyExists("pipe", pipe, err)
	}
	resourceID, err := resourceResult.LastInsertId()
	if err != nil {
		return err
	}

	if err := s.storeResourceLabels(ctx, tx, resourceID, resource.GetMetadata().GetLabels()); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *ResourceStore) StoreViewResource(ctx context.Context, resource *api.Resource, viewUUID string) error {
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

	resourceResult, err := tx.ExecContext(ctx, `INSERT INTO Resources (APIVersion, Kind, Version, Description, ViewID)
SELECT ?, ?, ?, ?, v.ID
FROM Views AS v
WHERE v.UUID = ?
LIMIT 1`,
		resource.GetAPIVersion(),
		resource.GetKind(),
		resource.GetMetadata().GetVersion(),
		resource.GetMetadata().GetDescription(),
		viewUUID,
	)
	if err != nil {
		// Currently Views get assigned a UUID so we don't expect conflicts here.
		return err
	}
	resourceID, err := resourceResult.LastInsertId()
	if err != nil {
		return err
	}

	if err := s.storeResourceLabels(ctx, tx, resourceID, resource.GetMetadata().GetLabels()); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *ResourceStore) storeResourceLabels(ctx context.Context, tx *sql.Tx, resourceID int64, labels []string) error {
	prep, err := tx.PrepareContext(ctx, "INSERT INTO ResourceLabels (ResourceID, Label) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer prep.Close()

	for _, label := range labels {
		if _, err := prep.ExecContext(ctx, resourceID, label); err != nil {
			return ignoreAlreadyExists(err)
		}
	}

	return nil
}
