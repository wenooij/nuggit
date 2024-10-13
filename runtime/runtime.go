package runtime

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/nodes"
	"github.com/wenooij/nuggit/status"
)

type Runtime struct {
	pipelines  map[string]*nuggit.Pipeline    // pipeline name => Pipeline.
	nodes      map[string]*nuggit.RawNode     // node name => Node.
	nodeUses   map[string]map[string]struct{} // node name => pipeline name => {}.
	conditions map[string]*runCond            // node name => cond.
	state      map[string]*nodeState          // node name => state.
	mu         sync.Mutex

	supportedActions  map[string]struct{}
	collections       map[string]struct{}
	pipelinesByHost   map[string][]*nuggit.Pipeline
	alwaysOnPipelines map[string]*nuggit.Pipeline
	dataIDs           map[nuggit.DataSpecifier]struct{}
}

func NewRuntime() (*Runtime, error) {
	// Add builtin pipeline.
	u, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	pipelines := map[string]*nuggit.Pipeline{
		"document": {Sequence: []string{u.String()}},
	}
	docNode, err := nuggit.Node[*nodes.Document]{Action: "document"}.Raw()
	if err != nil {
		return nil, err
	}
	docNodeID := u.String()
	nodes := map[string]*nuggit.RawNode{
		docNodeID: &docNode,
	}
	return &Runtime{
		pipelines:         pipelines,
		nodes:             nodes,
		nodeUses:          map[string]map[string]struct{}{docNodeID: {"document": {}}},
		supportedActions:  make(map[string]struct{}),
		collections:       make(map[string]struct{}),
		pipelinesByHost:   make(map[string][]*nuggit.Pipeline),
		alwaysOnPipelines: make(map[string]*nuggit.Pipeline),
		dataIDs:           make(map[nuggit.DataSpecifier]struct{}),
	}, nil
}

// locks excluded: mu.
func (r *Runtime) putNode(nodeID string, node *nuggit.RawNode) {
	if r.nodes[nodeID] = node; r.nodeUses[nodeID] == nil {
		r.nodeUses[nodeID] = map[string]struct{}{}
	}
}

// locks excluded: mu.
func (r *Runtime) putPipeNode(pipeID, nodeID string, node *nuggit.RawNode) {
	r.putNode(nodeID, node)
	set := r.nodeUses[nodeID]
	if set == nil {
		set = make(map[string]struct{}, 1)
		r.nodeUses[nodeID] = set
	}
	set[pipeID] = struct{}{}
}

// locks excluded: mu.
func (r *Runtime) deleteNode(nodeID string) bool {
	if len(r.nodeUses[nodeID]) != 0 {
		return false
	}
	delete(r.nodes, nodeID)
	delete(r.nodeUses, nodeID)
	return true
}

// locks excluded: mu.
func (r *Runtime) deletePipeNode(pipeID, nodeId string, keepNode bool) {
	if delete(r.nodeUses[nodeId], pipeID); !keepNode && len(r.nodeUses[nodeId]) == 0 {
		r.deleteNode(nodeId)
	}
}

// locks excluded: mu.
func (r *Runtime) deletePipeline(pipeID string, keepNodes bool) {
	pipe, ok := r.pipelines[pipeID]
	if !ok {
		return
	}
	for _, nodeID := range pipe.Sequence {
		r.deletePipeNode(pipeID, nodeID, keepNodes)
	}
	delete(r.pipelines, pipeID)
}

// locks excluded: mu.
func (r *Runtime) run(pipeline string) error {
	p := r.pipelines[pipeline]
	if p == nil {
		return fmt.Errorf("failed to get pipeline: %w", status.ErrNotFound)
	}
	return status.ErrUnimplemented
}
