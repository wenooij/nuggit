package main

import (
	"net/http"

	"github.com/wenooij/nuggit/runtime"
	"github.com/wenooij/nuggit/status"
)

func (a *API) RegisterPipelinesAPI() {
	a.HandleFunc("GET /api/pipelines/list", func(w http.ResponseWriter, r *http.Request) {
		resp, err := a.rt.ListPipelines(&runtime.ListPipelinesRequest{})
		status.WriteResponse(w, resp, err)
	})
	a.HandleFunc("GET /api/pipelines/{pipeline}", func(w http.ResponseWriter, r *http.Request) {
		resp, err := a.rt.ListPipeline(&runtime.ListPipelineRequest{Pipeline: r.PathValue("pipeline")})
		status.WriteResponse(w, resp, err)
	})
	a.HandleFunc("PUT /api/pipelines", func(w http.ResponseWriter, r *http.Request) {
		req := new(runtime.PutPipelineRequest)
		if !status.ReadRequest(w, r.Body, req) {
			return
		}
		resp, err := a.rt.PutPipeline(req)
		status.WriteResponse(w, resp, err)
	})
	a.HandleFunc("GET /api/pipelines/{pipeline}/status", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	a.HandleFunc("PATCH /api/pipelines/{pipeline}/status", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	a.HandleFunc("POST /api/pipelines/{pipeline}/trigger", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
}
