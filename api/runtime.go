package api

import (
	"fmt"
	"sync"
)

type RuntimeLite struct {
	*Ref `json:",omitempty"`
}

type RuntimeBase struct {
	Name             string        `json:"name,omitempty"`
	SupportedActions []*ActionLite `json:"supported_actions,omitempty"`
}

type Runtime struct {
	*RuntimeLite `json:",omitempty"`
	*RuntimeBase `json:",omitempty"`
}

type RuntimesAPI struct {
	rc        map[string]*Runtime
	defaultRc string
	mu        sync.RWMutex
}

func (a *RuntimesAPI) Init(storeType StorageType) error {
	*a = RuntimesAPI{
		rc: make(map[string]*Runtime),
	}
	rc, err := a.createRuntimeConfig("default")
	if err != nil {
		return err
	}
	a.defaultRc = rc.ID
	a.createRuntime(rc) // Always returns true.
	return nil
}

func (a *RuntimesAPI) createRuntimeConfig(name string) (*Runtime, error) {
	id, err := newUUID(func(id string) bool { return true })
	if err != nil {
		return nil, err
	}
	return &Runtime{
		RuntimeLite: &RuntimeLite{
			Ref: &Ref{
				ID:  id,
				URI: fmt.Sprintf("/api/runtimes/%s", id),
			},
		},
		RuntimeBase: &RuntimeBase{Name: name},
	}, nil
}

// locks excluded: mu.
func (a *RuntimesAPI) createRuntime(rt *Runtime) bool {
	if a.rc[rt.ID] != nil {
		return false
	}
	a.rc[rt.ID] = rt
	return true
}

type RuntimeStatusRequest struct{}

type RuntimeStatusResponse struct{}

func (a *RuntimesAPI) RuntimeStatus(*RuntimeStatusRequest) (*RuntimeStatusResponse, error) {
	return &RuntimeStatusResponse{}, nil
}
