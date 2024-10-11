package runtime

import (
	"github.com/wenooij/nuggit"
)

type Runtime struct {
	pipelines         map[Versioned[string]]*nuggit.Pipeline
	supportedActions  map[Versioned[string]]struct{}
	collections       map[Versioned[string]]struct{}
	pipelinesByHost   map[Versioned[string]][]*nuggit.Pipeline
	alwaysOnPipelines map[Versioned[string]]*nuggit.Pipeline
	dataIDs           map[Versioned[nuggit.DataSpecifier]]struct{}
}

func NewRuntime() *Runtime {
	return &Runtime{
		pipelines:         make(map[Versioned[string]]*nuggit.Pipeline),
		supportedActions:  make(map[Versioned[string]]struct{}),
		collections:       make(map[Versioned[string]]struct{}),
		pipelinesByHost:   make(map[Versioned[string]][]*nuggit.Pipeline),
		alwaysOnPipelines: make(map[Versioned[string]]*nuggit.Pipeline),
		dataIDs:           make(map[Versioned[nuggit.DataSpecifier]]struct{}),
	}
}

func Run(pipeline string) {

}
