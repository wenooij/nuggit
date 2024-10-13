package main

import (
	"net/http"

	"github.com/wenooij/nuggit/status"
)

func (a *API) RegisterCollectionsAPI() {
	a.HandleFunc("GET /api/collections/list", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	a.HandleFunc("GET /api/collections/{collection}", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	a.HandleFunc("DELETE /api/collections/{collection}", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	a.HandleFunc("GET /api/collections/{collection}/point/{name}", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	a.HandleFunc("DELETE /api/collections/{collection}/point/{name}", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	a.HandleFunc("GET /api/collections/{collection}/point/{name}/list", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
}
