package runtime

import (
	"fmt"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/client"
)

type Pipeline struct {
	Pipeline nuggit.Pipeline
	Disabled bool
}

type RunRequest struct {
	Args client.Args
	Data []byte
}

type RunResponse struct{}

func (r *Runtime) Run(*RunRequest) (*RunResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

type RunRequestBatch struct {
	Args []client.Args
}

type RunResponseBatch struct{}

func (r *Runtime) RunBatch(*RunRequestBatch) (*RunResponseBatch, error) {
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

type Error struct {
	Err        error
	StatusCode int
}

type ListPipelinesRequest struct{}

type ListPipelinesResponse struct {
	Pipelines map[string]nuggit.Pipeline `json:"pipelines,omitempty"`
}

func (r *Runtime) ListPipelines(*ListPipelinesRequest) (*ListPipelinesResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	res := make(map[string]nuggit.Pipeline, len(r.pipelines))
	for id, p := range r.pipelines {
		res[id] = *p
	}
	return &ListPipelinesResponse{Pipelines: res}, nil
}
