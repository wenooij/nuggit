package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/wenooij/nuggit/status"
)

type NodeLite struct {
	*Ref `json:",omitempty"`
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

type NodesAPI struct {
	api      *API
	pipes    *PipesAPI
	nodes    map[string]*Node               // node ID => Node.
	nodeDeps map[string]map[string]struct{} // pipe dependencies; node ID => pipe ID => {}.
	mu       sync.RWMutex
}

func (a *NodesAPI) Init(api *API, pipes *PipesAPI) {
	*a = NodesAPI{
		api:      api,
		pipes:    pipes,
		nodes:    make(map[string]*Node),
		nodeDeps: make(map[string]map[string]struct{}),
	}
}

// locks excluded: mu.
func (a *NodesAPI) deleteNode(nodeID string) bool {
	if len(a.nodeDeps[nodeID]) != 0 {
		return false
	}
	delete(a.nodes, nodeID)
	delete(a.nodeDeps, nodeID)
	return true
}

// locks excluded: mu.
func (a *NodesAPI) deletePipeNode(pipeID, nodeId string, keepNode bool) {
	if delete(a.nodeDeps[nodeId], pipeID); !keepNode && len(a.nodeDeps[nodeId]) == 0 {
		a.deleteNode(nodeId)
	}
}

// locks excluded: mu.
func (a *NodesAPI) createNode(node *Node) bool {
	if a.nodes[node.ID] != nil {
		return false
	}
	a.nodes[node.ID] = node
	a.nodeDeps[node.ID] = map[string]struct{}{}
	return true
}

// locks excluded: mu.
func (a *NodesAPI) createPipeNode(pipe *PipeLite, node *Node) bool {
	if !a.createNode(node) {
		return false
	}
	set := a.nodeDeps[node.ID]
	if set == nil {
		set = make(map[string]struct{}, 1)
		a.nodeDeps[node.ID] = set
	}
	set[pipe.ID] = struct{}{}
	return true
}

type ListNodesRequest struct{}

type ListNodesResponse struct {
	Nodes []*NodeLite `json:"nodes,omitempty"`
}

func (a *NodesAPI) ListNodes(*ListNodesRequest) (*ListNodesResponse, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	res := make([]*NodeLite, 0, len(a.nodes))
	for _, node := range a.nodes {
		res = append(res, node.NodeLite)
	}
	return &ListNodesResponse{Nodes: res}, nil
}

type GetNodeRequest struct {
	ID string `json:"id,omitempty"`
}

type GetNodeResponse struct {
	Node *Node `json:"node,omitempty"`
}

func (a *NodesAPI) GetNode(req *GetNodeRequest) (*GetNodeResponse, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	node, ok := a.nodes[req.ID]
	if !ok {
		return nil, fmt.Errorf("failed to load node: %w", status.ErrNotFound)
	}
	return &GetNodeResponse{Node: node}, nil
}

type GetNodesBatchRequest struct {
	Nodes []*NodeLite `json:"id,omitempty"`
}

type GetNodesBatchResponse struct {
	Nodes        []*Node  `json:"node,omitempty"`
	MissingNodes []string `json:"missing_nodes,omitempty"`
}

func (a *NodesAPI) GetNodesBatch(req *GetNodesBatchRequest) (*GetNodesBatchResponse, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	nodes := make([]*Node, 0, len(req.Nodes))
	var missingNodes []string
	for _, nl := range req.Nodes {
		if node := a.nodes[nl.ID]; node != nil {
			nodes = append(nodes, node)
		} else {
			missingNodes = append(missingNodes, nl.ID)
		}
	}
	return &GetNodesBatchResponse{Nodes: nodes, MissingNodes: missingNodes}, nil
}

type DeleteNodeRequest struct {
	ID string `json:"id,omitempty"`
}

type DeleteNodeResponse struct{}

func (r *NodesAPI) DeleteNode(req *DeleteNodeRequest) (*DeleteNodeResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := provided("id", req.ID); err != nil {
		return nil, err
	}
	if err := validateUUID(req.ID); err != nil {
		return nil, err
	}
	if r.nodes[req.ID] != nil && !r.deleteNode(req.ID) {
		return nil, fmt.Errorf("node is in use by at least one pipe: %w", status.ErrFailedPrecondition)
	}
	return &DeleteNodeResponse{}, nil
}

type CreateNodeRequest struct {
	Node *NodeBase `json:"node,omitempty"`
}

type CreateNodeResponse struct {
	Node *NodeLite `json:"node,omitempty"`
}

func (a *NodesAPI) CreateNode(req *CreateNodeRequest) (*CreateNodeResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := provided("node", req.Node); err != nil {
		return nil, err
	}
	id, err := newUUID(func(id string) bool { return a.nodes[id] == nil })
	if err != nil {
		return nil, err
	}
	node := &Node{
		NodeLite: &NodeLite{
			Ref: &Ref{
				ID:  id,
				URI: fmt.Sprintf("/api/nodes/%s", id),
			},
		},
	}
	if !a.createNode(node) {
		return nil, fmt.Errorf("failed to create node: %w", status.ErrAlreadyExists)
	}
	return &CreateNodeResponse{Node: node.NodeLite}, nil
}

type ListOrphansRequest struct{}

type ListOrphansResponse struct {
	Nodes []*NodeLite `json:"nodes,omitempty"`
}

func (r *NodesAPI) ListOrphans(*ListOrphansRequest) (*ListOrphansResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	orphans := make([]*NodeLite, 0, len(r.nodes))
	// Every node retains an entry in node deps.
	for id, deps := range r.nodeDeps {
		if len(deps) == 0 {
			orphans = append(orphans, r.nodes[id].NodeLite)
		}
	}
	return &ListOrphansResponse{Nodes: orphans}, nil
}

type DeleteOrphansRequest struct{}

type DeleteOrphansResponse struct{}

func (a *NodesAPI) DeleteOrphans(*DeleteOrphansRequest) (*DeleteOrphansResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	orphans := make([]string, 0, len(a.nodes))
	for id, deps := range a.nodeDeps {
		if len(deps) == 0 {
			orphans = append(orphans, id)
		}
	}
	for _, id := range orphans {
		a.deleteNode(id) // deleteNode always returns true because id is an orphan.
	}
	return &DeleteOrphansResponse{}, nil
}

type GetNodeDependenciesRequest struct {
	ID string `json:"id,omitempty"`
}

type GetNodeDependenciesResponse struct {
	Dependencies []*PipeLite `json:"dependencies,omitempty"`
}

func (a *NodesAPI) GetNodeDependencies(req *GetNodeDependenciesRequest) (*GetNodeDependenciesResponse, error) {
	a.api.mu.Lock()
	defer a.api.mu.Lock()
	a.mu.RLock()
	defer a.mu.RUnlock()
	a.pipes.mu.RLock()
	defer a.pipes.mu.RUnlock()

	pipes, ok := a.nodeDeps[req.ID]
	if !ok {
		return nil, fmt.Errorf("failed to load node: %w", status.ErrNotFound)
	}
	resp := make([]*PipeLite, 0, len(pipes))
	for id := range pipes {
		resp = append(resp, a.pipes.pipes[id].PipeLite)
	}
	return &GetNodeDependenciesResponse{Dependencies: resp}, nil
}
