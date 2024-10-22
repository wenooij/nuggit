package storage

import (
	"context"
	"database/sql"
	"iter"
	"log"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/table"
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

	log.Println("InsertRow() prepared: \n", insertQuery)

	pipesPoints := make(map[api.NameDigest]*api.Point)
	for _, p := range pipes {
		pipesPoints[p.NameDigest] = p.GetPoint()
	}

	// Flat unmarshal the args.
	// Each arg is a pull-style iterator.
	var flatArgs []nextStop
	for _, r := range row {
		point := pipesPoints[r.GetPipe()]
		data := r.Result
		next, stop := iter.Pull2(point.UnmarshalFlat(data))
		flatArgs = append(flatArgs, nextStop{next, stop})
	}

	// Continue inserting rows until the first iter is drained.
	for {
		args := make([]any, len(flatArgs))
		var stop bool
		for i, it := range flatArgs {
			v, err, ok := it.next()
			if err != nil {
				return err
			}
			if !ok {
				stop = true
				break
			}
			args[i] = v
		}
		if stop {
			for _, it := range flatArgs {
				it.stop()
			}
			break
		}
		if _, err := prep.ExecContext(ctx, args...); err != nil {
			return err
		}
	}

	return nil
}
