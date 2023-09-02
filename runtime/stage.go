package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"sync"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/edges"
	"github.com/wenooij/nuggit/graphs"
	"github.com/wenooij/nuggit/jsong"
	"github.com/wenooij/nuggit/nodes"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// MaxWorkers defines the maximum Node-level concurrency in the StageRunner.
var MaxWorkers = runtime.GOMAXPROCS(0)

// StageRunner executes all the graphs in the current stage concurrently.
// Multiple graphs in a given stage are executed in parallel while
// Node-level paralelism is limited by MaxWorkers.
//
// The StageCoordinator is used when the runner must exchange data
// outside the stage.
//
// The zero StageRunner is ready to use.
//
// TODO(wes): Implement the coordinator.
type StageRunner struct {
	NodeFactory NodeFactory
	Overrides   []map[string]any
	Coord       *StageCoordinator
	Graphs      []*graphs.Graph
	Results     []map[nuggit.NodeKey]json.RawMessage
	sem         *semaphore.Weighted
	once        sync.Once
}

func (r *StageRunner) initOnce() {
	r.sem = semaphore.NewWeighted(int64(MaxWorkers))
}

func (r *StageRunner) override(g *graphs.Graph, k nuggit.NodeKey, val any) error {
	n, ok := g.Nodes[k]
	if !ok {
		return fmt.Errorf("override: node with key not found: %q", k)
	}
	data, err := jsong.Merge(n.Data, val, "", "")
	if err != nil {
		return err
	}
	n.Data = data
	g.Nodes[k] = n
	return nil
}

func (r *StageRunner) Run(ctx context.Context) error {
	r.once.Do(r.initOnce)
	r.Results = make([]map[nuggit.NodeKey]json.RawMessage, len(r.Graphs))

	var eg errgroup.Group

	for i := range r.Graphs {
		i := i
		eg.Go(func() error { return r.runGraph(ctx, i) })
	}

	return eg.Wait()
}

func (r *StageRunner) runGraph(ctx context.Context, i int) error {
	g := r.Graphs[i]
	for k, val := range r.Overrides[i] {
		r.override(g, k, val)
	}
	results := make(map[nuggit.NodeKey]json.RawMessage, len(g.Nodes))

	ns := maps.Values(g.Nodes)
	ks := nodes.Keys(ns)
	if err := g.Sort(ks); err != nil {
		return fmt.Errorf("StageRunner: failed while preparing graph topology: %v", err)
	}

	for _, k := range ks {
		n := g.Nodes[k]
		op, err := r.NodeFactory.New(n)
		if err != nil {
			return fmt.Errorf("failed calling runner factory for node %v(%v): %v", n.Op, k, err)
		}
		err = func() error {
			if err := r.sem.Acquire(ctx, 1); err != nil {
				return fmt.Errorf("failed while waiting on resources for node %v(%v): %v", n.Op, k, err)
			}
			defer r.sem.Release(1)
			log.Printf("Binding %v(%v)", n.Op, k)

			data := n.Data
			for _, e := range g.Adjacency[k].Edges {
				e := g.Edges[e]
				log.Printf("  %s", edges.Format(e))
				data, err = jsong.Merge(data,
					results[e.Dst],
					e.SrcField,
					e.DstField,
				)
				if err != nil {
					return fmt.Errorf("failed while binding edge %v(%v).%q: %v", n.Op, k, e.Key, err)
				}
				log.Printf("Merged %v(%v): %s", n.Op, k, data)
			}

			res := op
			if runner, ok := op.(Runner); ok {
				log.Printf("Running %v(%v)", n.Op, k)
				res, err = func() (res any, err error) {
					defer func() {
						if rv := recover(); rv != nil {
							err = fmt.Errorf("recovered from panic: %v", rv)
						}
					}()
					return runner.Run(ctx)
				}()
				if err != nil {
					return fmt.Errorf("failed while executing node %v(%v): %v", n.Op, k, err)
				}
			}

			resultData, err := json.Marshal(res)
			if err != nil {
				return fmt.Errorf("failed while marshaling result for node: %v(%v): %w", n.Op, k, err)
			}

			if res != op {
				log.Printf("Finished %v(%v): %s", n.Op, k, string(resultData))
			}

			results[k] = resultData
			return nil
		}()
		if err != nil {
			return fmt.Errorf("StageRunner: %w", err)
		}
	}

	r.Results[i] = results
	return nil
}
