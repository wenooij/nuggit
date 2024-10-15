package api

import (
	"bytes"
	"encoding/json"
	"errors"
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
	api     *API
	pipes   *PipesAPI
	storage StoreInterface[*NodeRich]
	mu      sync.RWMutex
}

func (a *NodesAPI) Init(api *API, pipes *PipesAPI, storeType StorageType) error {
	*a = NodesAPI{
		api:   api,
		pipes: pipes,
	}
	switch storeType {
	case StorageInMemory:
		a.storage = newStorageInMemory[*NodeRich]()
		return nil
	default:
		return fmt.Errorf("persistent node storage is not supported: %w", status.ErrFailedPrecondition)
	}
}

// locks excluded: mu.
func (a *NodesAPI) loadNode(id string) (*NodeRich, error) { return a.storage.Load(id) }

// locks excluded: mu.
func (a *NodesAPI) deleteNode(nodeID string) (*StorageOpLite, error) {
	node, err := a.loadNode(nodeID)
	if err != nil {
		if errors.Is(err, status.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if len(node.Dependencies) != 0 {
		return nil, fmt.Errorf("node is in use by at least one pipe: %w", status.ErrFailedPrecondition)
	}
	return a.storage.Delete(nodeID)
}

// locks excluded: mu.
func (a *NodesAPI) deletePipeNode(pipeID, nodeID string, keepNode bool) (*StorageOpLite, error) {
	node, err := a.loadNode(nodeID)
	if err != nil {
		if errors.Is(err, status.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if delete(node.Dependencies, pipeID); !keepNode && len(node.Dependencies) == 0 {
		return a.deleteNode(nodeID)
	} else {
		return a.storage.StoreOrReplace(node)
	}
}

// locks excluded: mu.
func (a *NodesAPI) createNode(node *NodeRich) (*StorageOpLite, error) { return a.storage.Store(node) }

// locks excluded: mu.
func (a *NodesAPI) createPipeNode(pipe *PipeLite, node *NodeRich) (*StorageOpLite, error) {
	node.Dependencies[pipe.ID] = struct{}{}
	return a.createNode(node)
}

type ListNodesRequest struct{}

type ListNodesResponse struct {
	Nodes []*NodeLite `json:"nodes,omitempty"`
}

func (a *NodesAPI) ListNodes(*ListNodesRequest) (*ListNodesResponse, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	n, _ := a.storage.Len()
	res := make([]*NodeLite, 0, n)
	a.storage.Scan(func(n *NodeRich, err error) error {
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

func (a *NodesAPI) GetNode(req *GetNodeRequest) (*GetNodeResponse, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	node, err := a.storage.Load(req.ID)
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

func (a *NodesAPI) GetNodesBatch(req *GetNodesBatchRequest) (*GetNodesBatchResponse, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	nodes := make([]*NodeRich, 0, len(req.Nodes))
	var missingNodes []string
	for _, nl := range req.Nodes {
		if node, err := a.storage.Load(nl.ID); err == nil {
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

func (a *NodesAPI) DeleteNode(req *DeleteNodeRequest) (*DeleteNodeResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := provided("id", req.ID); err != nil {
		return nil, err
	}
	if err := validateUUID(req.ID); err != nil {
		return nil, err
	}
	if _, err := a.deleteNode(req.ID); err != nil {
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

func (a *NodesAPI) CreateNode(req *CreateNodeRequest) (*CreateNodeResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := provided("node", req.Node); err != nil {
		return nil, err
	}
	id, err := newUUID(func(id string) bool { _, err := a.loadNode(id); return errors.Is(err, status.ErrNotFound) })
	if err != nil {
		return nil, err
	}
	node := &NodeRich{
		Node: &Node{
			NodeLite: &NodeLite{
				Ref: &Ref{
					ID:  id,
					URI: fmt.Sprintf("/api/nodes/%s", id),
				},
			},
			NodeBase: req.Node,
		},
		NodeState: &NodeState{},
	}
	if _, err := a.createNode(node); err != nil {
		return nil, err
	}
	return &CreateNodeResponse{Node: node.NodeLite}, nil
}

type ListOrphansRequest struct{}

type ListOrphansResponse struct {
	Nodes []*NodeLite `json:"nodes,omitempty"`
}

func (a *NodesAPI) ListOrphans(*ListOrphansRequest) (*ListOrphansResponse, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	orphans := make([]*NodeLite, 0, 64)
	a.storage.Scan(func(n *NodeRich, err error) error {
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

func (a *NodesAPI) DeleteOrphans(*DeleteOrphansRequest) (*DeleteOrphansResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	orphans := make([]string, 0, 64)
	a.storage.Scan(func(n *NodeRich, err error) error {
		if err != nil {
			return err
		}
		if len(n.GetState().Dependencies) == 0 {
			orphans = append(orphans, n.ID)
		}
		return nil
	})
	for _, id := range orphans {
		a.deleteNode(id) // deleteNode should always succeed.
	}
	return &DeleteOrphansResponse{}, nil
}
