package runtime

import (
	"fmt"

	"github.com/wenooij/nuggit"
)

type Versioned[E any] struct {
	Elem    E
	Version string
}

type Pipeline struct {
	Pipeline nuggit.Pipeline
	Disabled bool
}

type RunRequest struct {
	URL  string
	Data []byte
}

type RunResponse struct{}

func (r *Runtime) Run(*RunRequest) (*RunResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

type EnableRequest struct {
	Name    string
	Enabled bool
}

type EnableResponse struct{}

func (r *Runtime) Enable(*EnableRequest) (*EnableResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

type DeletePipelineRequest struct {
	Name string
}

type DeletePipelineResponse struct{}

func (r *Runtime) Delete(*DeletePipelineRequest) (*DeletePipelineResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

type DeletePipelineRequestBatch struct {
	Names []string
}

type DeletePipelineResponseBatch struct{}

func (r *Runtime) DeleteBatch(*DeletePipelineRequestBatch) (*DeletePipelineResponseBatch, error) {
	return nil, fmt.Errorf("not implemented")
}

type ReplacePipelineRequest struct {
	Name     string
	Pipeline nuggit.Pipeline
}

type ReplacePipelineResponse struct {
	Name     string
	Pipeline nuggit.Pipeline
}

func (r *Runtime) Replace(*ReplacePipelineRequest) (*ReplacePipelineResponse, error) {
	return nil, fmt.Errorf("not implemented")
}
