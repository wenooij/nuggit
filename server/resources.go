package main

import (
	"net/http"

	"github.com/wenooij/nuggit/status"
)

func (a *API) RegisterResourcesAPI() {
	a.HandleFunc("GET /api/resources/list", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	a.HandleFunc("GET /api/resources/versions/list", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	a.HandleFunc("PATCH /api/resources", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	a.HandleFunc("POST /api/resources", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	a.HandleFunc("PUT /api/resources", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	a.HandleFunc("GET /api/resources/{resource}", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	a.HandleFunc("DELETE /api/resources/{resource}", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
}
