package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/wenooij/nuggit/runtime"
	"github.com/wenooij/nuggit/status"
)

type Server struct{}

func main() {
	port := flag.Int("port", 9402, "Server port")
	nuggitDir := flag.String("nuggit_dir", filepath.Join(os.Getenv("HOME"), ".nuggit"), "Location of the Nuggit directory")

	info, err := os.Stat(*nuggitDir)
	if err != nil {
		log.Printf("Failed to access nuggit directory: %v", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		log.Printf("Nuggit path is not a directory: %v", *nuggitDir)
		os.Exit(2)
	}
	rt := runtime.NewRuntime()

	http.HandleFunc("GET /api/status", func(w http.ResponseWriter, r *http.Request) { status.WriteResponse(w, struct{}{}, nil) })
	http.HandleFunc("GET /api/pipelines/list", func(w http.ResponseWriter, r *http.Request) {
		resp, err := rt.ListPipelines(&runtime.ListPipelinesRequest{})
		status.WriteResponse(w, resp, err)
	})
	http.HandleFunc("PATCH /api/pipelines/{pipeline}/create", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	http.HandleFunc("GET /api/pipelines/{pipeline}/status", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	http.HandleFunc("POST /api/pipelines/{pipeline}/run", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	http.HandleFunc("PATCH /api/pipelines/{pipeline}/enable", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	http.HandleFunc("GET /api/runtime/status", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	http.HandleFunc("GET /api/runtime/{batch}/results/list", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	http.HandleFunc("GET /api/runtime/{batch}/results/{result}/list", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	http.HandleFunc("GET /api/runtime/{batch}/status", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	http.HandleFunc("GET /api/collections/list", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	http.HandleFunc("GET /api/collections/{collection}", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	http.HandleFunc("DELETE /api/collections/{collection}", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	http.HandleFunc("GET /api/collections/{collection}/data/{name}/list", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	http.ListenAndServe(fmt.Sprint(":", *port), http.DefaultServeMux)
}
