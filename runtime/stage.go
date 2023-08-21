package runtime

import (
	"context"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/graphs"
)

// StageRunner executes the given Keys of the graph in serial using a left-most DFS of the DAG.
// Keys outputted outside of the stage are sent to the stage coordinator.
//
// TODO(wes): Implement the coordinator.
type StageRunner struct {
	*graphs.Subgraph
	Factory RunnerFactory
	Coord   *StageCoordinator
}

func NewStageRunner(g *graphs.Graph, coord *StageCoordinator, rf RunnerFactory, k nuggit.StageKey) *StageRunner {
	return &StageRunner{
		Subgraph: g.StageGraph(k),
		Coord:    coord,
		Factory:  rf,
	}
}

type stageStackEntry struct {
	key        nuggit.NodeKey
	edgeOffset int
}

func (r *StageRunner) Run(ctx context.Context) error {
	visited := make(map[nuggit.NodeKey]struct{}, len(r.Nodes))
	stack := make([]stageStackEntry, 0, len(r.Nodes))
	results := make(map[nuggit.NodeKey]any, len(r.Nodes))

	for _, n := range r.Nodes {
		stack = append(stack, stageStackEntry{key: n.Key})
	}

	for len(stack) > 0 {
		n := len(stack) - 1
		e := stack[n]
		stack = stack[:n]
		if _, ok := visited[e.key]; ok {
			continue
		}
		visited[e.key] = struct{}{}

		a := r.Adjacency[e.key]
		es := a.Edges[e.edgeOffset:]

		if len(es) == 0 {
			re, err := r.Factory.NewRunner(r.Nodes[e.key])
			if err != nil {
				return err
			}
			es := make([]Edge, 0, len(r.Adjacency))
			for _, e := range r.Adjacency[e.key].Edges {
				edge := r.Edges[e]
				es = append(es, Edge{
					Edge:   edge,
					Result: results[edge.Dst],
				})
			}
			if binder, ok := re.(Binder); ok {
				if err := binder.Bind(es); err != nil {
					return err
				}
			}
			res, err := re.Run(ctx)
			if err != nil {
				return err
			}
			results[e.key] = res
			continue
		}

		stack = append(stack,
			stageStackEntry{key: e.key, edgeOffset: e.edgeOffset + 1},
			stageStackEntry{key: es[0]},
		)
	}

	return nil
}
