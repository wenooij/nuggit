package main

import (
	"net/http"

	"github.com/wenooij/nuggit/runtime"
	"github.com/wenooij/nuggit/status"
)

func (a *API) RegisterRuntimeAPI() {
	a.HandleFunc("GET /api/runtime/status", func(w http.ResponseWriter, r *http.Request) {
		resp, err := a.rt.Status(&runtime.StatusRequest{})
		status.WriteResponse(w, resp, err)
	})
	a.HandleFunc("GET /api/runtime/stats", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	a.HandleFunc("GET /api/runtime/list", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
}
