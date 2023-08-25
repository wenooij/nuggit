package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"sync"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/graphs"
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
	Factory RunnerFactory
	Coord   *StageCoordinator
	Graphs  []*nuggit.Graph
	Results []map[nuggit.NodeKey]any
	sem     *semaphore.Weighted
	once    sync.Once
}

func (r *StageRunner) initOnce() {
	r.sem = semaphore.NewWeighted(int64(MaxWorkers))
}

func (r *StageRunner) Run(ctx context.Context) error {
	r.once.Do(r.initOnce)
	r.Results = make([]map[nuggit.NodeKey]any, len(r.Graphs))

	var eg errgroup.Group

	for i := range r.Graphs {
		i := i
		eg.Go(func() error { return r.runGraph(ctx, i) })
	}

	return eg.Wait()
}

func (r *StageRunner) runGraph(ctx context.Context, i int) error {
	g := graphs.FromGraph(r.Graphs[i])
	results := make(map[nuggit.NodeKey]any, len(g.Nodes))

	ns := maps.Values(g.Nodes)
	ks := nodes.Keys(ns)
	if err := g.Sort(ks); err != nil {
		return fmt.Errorf("StageRunner: failed while preparing graph topology: %v", err)
	}

	for _, k := range ks {
		n := g.Nodes[k]
		re, err := r.Factory.NewRunner(n)
		if err != nil {
			return fmt.Errorf("failed calling runner factory for node %v(%v): %v", n.Op, k, err)
		}
		err = func() error {
			if err := r.sem.Acquire(ctx, 1); err != nil {
				return fmt.Errorf("failed while waiting on resources for node %v(%v): %v", n.Op, k, err)
			}
			defer r.sem.Release(1)

			data, _ := json.MarshalIndent(n, "", "  ")
			log.Println("Starting node:", k, ":\n", string(data))

			if binder, ok := re.(Binder); ok {
				for _, e := range g.Adjacency[k].Edges {
					edge := g.Edges[e]
					e := Edge{
						Edge:   edge,
						Result: results[edge.Dst],
					}
					if err := func() (err error) {
						defer func() {
							if rv := recover(); rv != nil {
								err = fmt.Errorf("recovered from panic: %v", rv)
							}
						}()
						return binder.Bind(e)
					}(); err != nil {
						return fmt.Errorf("failed while binding edge %v(%v).%q: %v", n.Op, k, e.Key, err)
					}
				}
			}
			res, err := func() (res any, err error) {
				defer func() {
					if rv := recover(); rv != nil {
						err = fmt.Errorf("recovered from panic: %v", rv)
					}
				}()
				return re.Run(ctx)
			}()
			if err != nil {
				return fmt.Errorf("failed while executing node %v(%v): %v", n.Op, k, err)
			}
			results[k] = res
			return nil
		}()
		if err != nil {
			return fmt.Errorf("StageRunner: %w", err)
		}
	}

	r.Results[i] = results
	return nil
}
