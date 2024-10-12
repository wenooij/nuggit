package runtime

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/status"
)

type Runtime struct {
	pipelines map[string]*nuggit.Pipeline // name => version => pipeline
	mu        sync.Mutex

	supportedActions  map[string]struct{}
	collections       map[string]struct{}
	pipelinesByHost   map[string][]*nuggit.Pipeline
	alwaysOnPipelines map[string]*nuggit.Pipeline
	dataIDs           map[nuggit.DataSpecifier]struct{}
}

func NewRuntime() *Runtime {
	// Add builtin pipeline.
	pipelines := map[string]*nuggit.Pipeline{
		"document": {
			Ops: []nuggit.RawOp{{
				Action: "document",
				Spec:   json.RawMessage(`{}`),
			}},
		},
	}
	return &Runtime{
		pipelines:         pipelines,
		supportedActions:  make(map[string]struct{}),
		collections:       make(map[string]struct{}),
		pipelinesByHost:   make(map[string][]*nuggit.Pipeline),
		alwaysOnPipelines: make(map[string]*nuggit.Pipeline),
		dataIDs:           make(map[nuggit.DataSpecifier]struct{}),
	}
}

func (r *Runtime) run(pipeline string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p := r.pipelines[pipeline]
	if p == nil {
		return fmt.Errorf("failed to get pipeline: %w", status.ErrNotFound)
	}
	return status.ErrUnimplemented
}
