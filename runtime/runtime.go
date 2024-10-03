package runtime

import (
	"github.com/wenooij/nuggit"
)

type Runtime struct {
	pipelines         map[string]*nuggit.Pipeline
	supportedActions  map[string]struct{}
	collections       map[string]struct{}
	pipelinesByHost   map[string][]*nuggit.Pipeline
	alwaysOnPipelines map[string]*nuggit.Pipeline
	dataIDs           map[nuggit.DataSpecifier]struct{}
}

type Pipeline struct {
	Pipeline nuggit.Pipeline
	Disabled bool
}

func NewRuntime() *Runtime {
	return &Runtime{
		pipelines:         make(map[string]*nuggit.Pipeline),
		supportedActions:  make(map[string]struct{}),
		collections:       make(map[string]struct{}),
		pipelinesByHost:   make(map[string][]*nuggit.Pipeline),
		alwaysOnPipelines: make(map[string]*nuggit.Pipeline),
		dataIDs:           make(map[nuggit.DataSpecifier]struct{}),
	}
}

type RunRequest struct {
	URL  string
	Data []byte
}

type EnableRequest struct {
	Name    string
	Enabled bool
}

type RemovePipelineRequest struct {
	Name string
}

type RemovePipelineRequestBatch struct {
	Names []string
}

type ReplacePipelineRequest struct {
	Name     string
	Pipeline nuggit.Pipeline
}
