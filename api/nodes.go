package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/wenooij/nuggit/status"
)

type NodeLite struct {
	*Ref `json:",omitempty"`
}

func NewNodeLite(id string) *NodeLite {
	return &NodeLite{newRef("/api/nodes/", id)}
}

type NodeBase struct {
	Action string          `json:"action,omitempty"`
	Spec   json.RawMessage `json:"spec,omitempty"`
}

func (a *NodeBase) sameNode(b *NodeBase) bool {
	return a == nil && b == nil ||
		a != nil && b != nil && a.Action == b.Action && bytes.Equal(a.Spec, b.Spec)
}

type Node struct {
	*NodeLite `json:",omitempty"`
	*NodeBase `json:",omitempty"`
}

type NodeState struct {
	Dependencies map[string]struct{} `json:"dependencies,omitempty"`
}

type NodeRich struct {
	*Node      `json:",omitempty"`
	*NodeState `json:"state,omitempty"`
}

func (n *NodeRich) GetState() *NodeState {
	if n == nil || n.NodeState == nil {
		return &NodeState{}
	}
	return n.NodeState
}

type NodesAPI struct {
	pipes *PipesAPI
	store NodeStore
}

func (a *NodesAPI) Init(store NodeStore, pipes *PipesAPI) error {
	*a = NodesAPI{
		pipes: pipes,
		store: store,
	}
	return nil
}

// locks excluded: mu.
func (a *NodesAPI) loadNode(ctx context.Context, id string) (*NodeRich, error) {
	return a.store.Load(ctx, id)
}

// locks excluded: mu.
func (a *NodesAPI) deleteNode(ctx context.Context, nodeID string) error {
	node, err := a.loadNode(ctx, nodeID)
	if err != nil {
		if errors.Is(err, status.ErrNotFound) {
			return nil
		}
		return err
	}
	if len(node.Dependencies) != 0 {
		return fmt.Errorf("node is in use by at least one pipe: %w", status.ErrFailedPrecondition)
	}
	return a.store.Delete(ctx, nodeID)
}

// locks excluded: mu.
func (a *NodesAPI) deletePipeNode(ctx context.Context, pipeID, nodeID string, keepNode bool) error {
	node, err := a.loadNode(ctx, nodeID)
	if err != nil {
		if errors.Is(err, status.ErrNotFound) {
			return nil
		}
		return err
	}
	if delete(node.Dependencies, pipeID); !keepNode && len(node.Dependencies) == 0 {
		return a.deleteNode(ctx, nodeID)
	} else {
		return a.store.StoreOrReplace(ctx, node)
	}
}

// locks excluded: mu.
func (a *NodesAPI) createNode(ctx context.Context, node *NodeRich) error {
	return a.store.Store(ctx, node)
}

// locks excluded: mu.
func (a *NodesAPI) createPipeNode(ctx context.Context, pipe *PipeLite, node *NodeRich) error {
	node.Dependencies[pipe.ID] = struct{}{}
	return a.createNode(ctx, node)
}

type ListNodesRequest struct{}

type ListNodesResponse struct {
	Nodes []*NodeLite `json:"nodes,omitempty"`
}

func (a *NodesAPI) ListNodes(ctx context.Context, req *ListNodesRequest) (*ListNodesResponse, error) {
	n, _ := a.store.Len(ctx)
	res := make([]*NodeLite, 0, n)
	a.store.Scan(ctx, func(n *NodeRich, err error) error {
		if err != nil {
			return err
		}
		res = append(res, n.NodeLite)
		return nil
	})
	return &ListNodesResponse{Nodes: res}, nil
}

type GetNodeRequest struct {
	ID string `json:"id,omitempty"`
}

type GetNodeResponse struct {
	Node *NodeRich `json:"node,omitempty"`
}

func (a *NodesAPI) GetNode(ctx context.Context, req *GetNodeRequest) (*GetNodeResponse, error) {
	node, err := a.store.Load(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	return &GetNodeResponse{Node: node}, nil
}

type GetNodesBatchRequest struct {
	Nodes []*NodeLite `json:"id,omitempty"`
}

type GetNodesBatchResponse struct {
	Nodes        []*NodeRich `json:"node,omitempty"`
	MissingNodes []string    `json:"missing_nodes,omitempty"`
}

func (a *NodesAPI) GetNodesBatch(ctx context.Context, req *GetNodesBatchRequest) (*GetNodesBatchResponse, error) {
	nodes := make([]*NodeRich, 0, len(req.Nodes))
	var missingNodes []string
	for _, nl := range req.Nodes {
		if node, err := a.store.Load(ctx, nl.ID); err == nil {
			nodes = append(nodes, node)
		} else if errors.Is(err, status.ErrNotFound) {
			missingNodes = append(missingNodes, nl.ID)
		} else {
			return nil, err
		}
	}
	return &GetNodesBatchResponse{Nodes: nodes, MissingNodes: missingNodes}, nil
}

type DeleteNodeRequest struct {
	ID string `json:"id,omitempty"`
}

type DeleteNodeResponse struct{}

func (a *NodesAPI) DeleteNode(ctx context.Context, req *DeleteNodeRequest) (*DeleteNodeResponse, error) {
	if err := provided("id", "is", req.ID); err != nil {
		return nil, err
	}
	if err := validateUUID(req.ID); err != nil {
		return nil, err
	}
	if err := a.deleteNode(ctx, req.ID); err != nil {
		return nil, err
	}
	return &DeleteNodeResponse{}, nil
}

type CreateNodeRequest struct {
	Node *NodeBase `json:"node,omitempty"`
}

type CreateNodeResponse struct {
	Node *NodeLite `json:"node,omitempty"`
}

func (a *NodesAPI) CreateNode(ctx context.Context, req *CreateNodeRequest) (*CreateNodeResponse, error) {
	if err := provided("node", "is", req.Node); err != nil {
		return nil, err
	}
	id, err := newUUID(func(id string) error { _, err := a.loadNode(ctx, id); return err })
	if err != nil {
		return nil, err
	}
	node := &NodeRich{
		Node: &Node{
			NodeLite: NewNodeLite(id),
			NodeBase: req.Node,
		},
		NodeState: &NodeState{},
	}
	if err := a.createNode(ctx, node); err != nil {
		return nil, err
	}
	return &CreateNodeResponse{Node: node.NodeLite}, nil
}

type ListOrphansRequest struct{}

type ListOrphansResponse struct {
	Nodes []*NodeLite `json:"nodes,omitempty"`
}

func (a *NodesAPI) ListOrphans(ctx context.Context, _ *ListOrphansRequest) (*ListOrphansResponse, error) {
	orphans := make([]*NodeLite, 0, 64)
	a.store.Scan(ctx, func(n *NodeRich, err error) error {
		if err != nil {
			return err
		}
		if len(n.GetState().Dependencies) == 0 {
			orphans = append(orphans, n.NodeLite)
		}
		return nil
	})
	return &ListOrphansResponse{Nodes: orphans}, nil
}

type DeleteOrphansRequest struct{}

type DeleteOrphansResponse struct{}

func (a *NodesAPI) DeleteOrphans(ctx context.Context, _ *DeleteOrphansRequest) (*DeleteOrphansResponse, error) {
	orphans := make([]string, 0, 64)
	a.store.Scan(ctx, func(n *NodeRich, err error) error {
		if err != nil {
			return err
		}
		if len(n.GetState().Dependencies) == 0 {
			orphans = append(orphans, n.ID)
		}
		return nil
	})
	for _, id := range orphans {
		a.deleteNode(ctx, id) // deleteNode should always succeed.
	}
	return &DeleteOrphansResponse{}, nil
}
