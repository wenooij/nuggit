package storage

import (
	"context"
	"database/sql"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/table"
)

type ResultStore struct {
	db *sql.DB
}

func NewResultStore(db *sql.DB) *ResultStore {
	return &ResultStore{db: db}
}

func (s *ResultStore) InsertRow(ctx context.Context, c *api.Collection, pipes []*api.Pipe, row []api.ExchangeResult) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	var ib table.InsertBuilder
	ib.Reset(c)
	ib.Add(pipes...)

	insertQuery, err := ib.Build()
	if err != nil {
		return nil
	}

	prep, err := conn.PrepareContext(ctx, insertQuery)
	if err != nil {
		return err
	}
	defer prep.Close()

	pipesPoints := make(map[api.NameDigest]*api.Point)
	for _, p := range pipes {
		pipesPoints[p.NameDigest] = p.GetPoint()
	}

	var args []any
	for _, r := range row {
		point := pipesPoints[r.GetPipe()]
		data := r.Result
		v, err := point.UnmarshalNew(data)
		if err != nil {
			return err
		}
		args = append(args, v)
	}

	if _, err := prep.ExecContext(ctx, args...); err != nil {
		return err
	}

	return nil
}
