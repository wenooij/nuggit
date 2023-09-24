package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/edges"
	"github.com/wenooij/nuggit/graphs"
	"github.com/wenooij/nuggit/jsong"
	"golang.org/x/sync/errgroup"
)

type GraphRunner struct {
	*StageRunner
	*graphs.Graph
	NodeOverrides map[nuggit.NodeKey]any
	NodeData      map[nuggit.NodeKey]any
	NodeResults   map[nuggit.NodeKey]any
}

// StageRunner executes all the graphs in the current stage concurrently.
// Multiple graphs in a given stage are executed in parallel while
// Node-level paralelism is limited by MaxWorkers.
//
// The StageCoordinator is used when the runner must exchange data
// outside the stage.
//
// TODO(wes): Implement the coordinator.
type StageRunner struct {
	*StageCoordinator
	OpFactory    OpFactory
	GraphRunners []*GraphRunner
}

func (r *StageRunner) Run(ctx context.Context) error {
	var eg errgroup.Group

	for _, r := range r.GraphRunners {
		r := r
		eg.Go(func() error { return r.Run(ctx) })
	}

	return eg.Wait()
}

func debugResultData(data any) {
	bs, err := json.MarshalIndent(data, "  ", "  ")
	if err != nil {
		log.Printf("Failed to debug node: %v", err)
		return
	}
	log.Printf("\n  %v", string(bs))
}

func (r *GraphRunner) Run(ctx context.Context) error {
	results := make(map[nuggit.NodeKey]any, len(r.Nodes))
	graphs.Visit(r.Graph, func(k nuggit.NodeKey) error {
		n := r.Nodes[k]

		log.Printf("Creating %v(%v)", n.Op, n.Key)
		op, err := r.OpFactory.New(n)
		if err != nil {
			return fmt.Errorf("failed to create op for node: %v(%v): %w", n.Op, n.Key, err)
		}
		results[k] = op
		debugResultData(op)

		log.Printf("Merging %v(%v)", n.Op, n.Key)
		v, err := jsong.Merge(op, n.Data, "", "")
		if err != nil {
			return fmt.Errorf("failed to merge node: %v(%v): %w", n.Op, n.Key, err)
		}
		op = v
		debugResultData(op)

		log.Printf("Binding %v(%v)", n.Op, n.Key)
		if err := r.bindNode(results, n); err != nil {
			return fmt.Errorf("failed to bind node: %v(%v): %w", n.Op, n.Key, err)
		}
		debugResultData(op)

		if override, ok := r.NodeOverrides[k]; ok {
			log.Printf("Overriding %v(%v)", n.Op, n.Key)
			if err := r.override(k, override); err != nil {
				return fmt.Errorf("failed to override op for node: %v(%v): %w", n.Op, n.Key, err)
			}
			debugResultData(op)
		}

		log.Printf("Starting %v(%v)", n.Op, n.Key)
		if err := r.runOp(ctx, results, n, op); err != nil {
			return fmt.Errorf("failed to execute op for node: %v(%v): %w", n.Op, n.Key, err)
		}

		return nil
	})
	r.NodeResults = results
	return nil
}

func (r *GraphRunner) override(k nuggit.NodeKey, val any) error {
	n, ok := r.Nodes[k]
	if !ok {
		return fmt.Errorf("override: node with key not found: %q", k)
	}
	data, err := jsong.Merge(n.Data, val, "", "")
	if err != nil {
		return err
	}
	n.Data = data
	r.Nodes[k] = n
	return nil
}

func (r *GraphRunner) bindNode(results map[nuggit.NodeKey]any, n nuggit.Node) error {
	es := r.Adjacency[n.Key].Edges
	if len(es) == 0 {
		return nil
	}

	data := n.Data
	for _, e := range es {
		e := r.Edges[e]
		log.Printf("  %s", edges.Format(e))
		var err error
		data, err = jsong.Merge(data,
			results[e.Dst],
			e.SrcField,
			e.DstField,
		)
		if err != nil {
			return err
		}
	}
	n.Data = data
	return nil
}

func (r *GraphRunner) runOp(ctx context.Context, result map[nuggit.NodeKey]any, n nuggit.Node, op any) (err error) {
	res := op
	defer func() {
		log.Printf("Finished %v(%v)", n.Op, n.Key)
		debugResultData(res)
		result[n.Key] = res
	}()

	runner, ok := op.(Runner)
	if !ok {
		return nil
	}

	defer func() {
		if rv := recover(); rv != nil {
			err = fmt.Errorf("recovered from panic: %v", rv)
		}
	}()
	if res, err = runner.Run(ctx); err != nil {
		return err
	}

	result[n.Key] = res
	return nil
}
