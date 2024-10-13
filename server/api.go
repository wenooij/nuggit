package main

import (
	"log"
	"net/http"
	"slices"

	"github.com/wenooij/nuggit/runtime"
	"github.com/wenooij/nuggit/status"
)

type API struct {
	rt       *runtime.Runtime
	mux      *http.ServeMux
	patterns []string
}

func (a *API) HandleFunc(pattern string, handler http.HandlerFunc) {
	a.patterns = append(a.patterns, pattern)
	a.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s => %s ...", r.Method, r.URL.String(), pattern)
		handler(wrappedResponseWriter{w}, r)
	})
}

type wrappedResponseWriter struct {
	http.ResponseWriter
}

func (w wrappedResponseWriter) WriteHeader(statusCode int) {
	log.Printf("... %d %s", statusCode, http.StatusText(statusCode))
	w.ResponseWriter.WriteHeader(statusCode)
}

func (a *API) RegisterAPI() {
	a.HandleFunc("GET /api/list", func(w http.ResponseWriter, r *http.Request) { status.WriteResponse(w, a.patterns, nil) })
	a.HandleFunc("GET /api/status", func(w http.ResponseWriter, r *http.Request) { status.WriteResponse(w, struct{}{}, nil) })
	a.RegisterCollectionsAPI()
	a.RegisterNodesAPI()
	a.RegisterPipelinesAPI()
	a.RegisterRuntimeAPI()
	slices.Sort(a.patterns)
}
