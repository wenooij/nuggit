package main

import (
	"net/http"

	"github.com/wenooij/nuggit/runtime"
	"github.com/wenooij/nuggit/status"
)

func (a *API) RegisterNodesAPI() {
	a.HandleFunc("GET /api/nodes/list", func(w http.ResponseWriter, r *http.Request) {
		resp, err := a.rt.ListNodes(&runtime.ListNodesRequest{})
		status.WriteResponse(w, resp, err)
	})
	a.HandleFunc("GET /api/nodes/{node}", func(w http.ResponseWriter, r *http.Request) {
		resp, err := a.rt.ListNode(&runtime.ListNodeRequest{ID: r.PathValue("node")})
		status.WriteResponse(w, resp, err)
	})
	a.HandleFunc("GET /api/nodes/{node}/uses", func(w http.ResponseWriter, r *http.Request) {
		resp, err := a.rt.ListNode(&runtime.ListNodeRequest{ID: r.PathValue("node")})
		status.WriteResponse(w, resp, err)
	})
	a.HandleFunc("DELETE /api/nodes/{node}", func(w http.ResponseWriter, r *http.Request) {
		resp, err := a.rt.DeleteNode(&runtime.DeleteNodeRequest{ID: r.PathValue("node")})
		status.WriteResponse(w, resp, err)
	})
	a.HandleFunc("PUT /api/nodes", func(w http.ResponseWriter, r *http.Request) {
		req := new(runtime.PutNodeRequest)
		if !status.ReadRequest(w, r.Body, req) {
			return
		}
		resp, err := a.rt.PutNode(req)
		status.WriteResponse(w, resp, err)
	})
	a.HandleFunc("GET /api/nodes/orphans", func(w http.ResponseWriter, r *http.Request) {
		resp, err := a.rt.ListOrphans(&runtime.ListOrphansRequest{})
		status.WriteResponse(w, resp, err)
	})
	a.HandleFunc("DELETE /api/nodes/orphans", func(w http.ResponseWriter, r *http.Request) {
		resp, err := a.rt.DeleteOrphans(&runtime.DeleteOrphansRequest{})
		status.WriteResponse(w, resp, err)
	})
}
