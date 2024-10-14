package storage

import (
	"fmt"
	"sync"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/status"
)

type InMemory struct {
	resources map[string]*api.ResourceBase
	pipes     map[string]*api.PipeBase
	nodes     map[string]*api.NodeBase
	mu        sync.RWMutex
}

func NewInMemory() *InMemory {
	return &InMemory{
		resources: make(map[string]*api.ResourceBase),
		pipes:     make(map[string]*api.PipeBase),
		nodes:     make(map[string]*api.NodeBase),
	}
}

func (m *InMemory) Resource(id string) *api.ResourceBase {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.resources[id]
}

func (m *InMemory) Pipe(id string) *api.PipeBase {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pipes[id]
}

func (m *InMemory) Node(id string) *api.NodeBase {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.nodes[id]
}

func (m *InMemory) StoreResource(r *api.Resource) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.resources[r.ID] != nil {
		return fmt.Errorf("failed to store node: %w", status.ErrAlreadyExists)
	}
	m.resources[r.ID] = r.ResourceBase
	return nil
}

func (m *InMemory) StorePipe(p *api.Pipe) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.pipes[p.ID] != nil {
		return fmt.Errorf("failed to store node: %w", status.ErrAlreadyExists)
	}
	m.pipes[p.ID] = p.PipeBase
	return nil
}

func (m *InMemory) StoreNode(n *api.Node) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.nodes[n.ID] != nil {
		return fmt.Errorf("failed to store node: %w", status.ErrAlreadyExists)
	}
	m.nodes[n.ID] = n.NodeBase
	return nil
}
