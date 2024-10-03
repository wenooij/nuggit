package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/wenooij/nuggit/runtime"
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
	r := runtime.NewRuntime()
	_ = r

	http.HandleFunc("GET /api/status", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`{}`)) })
	http.HandleFunc("GET /api/pipelines/list", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) })
	http.HandleFunc("PATCH /api/pipelines/{pipeline}/create", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) })
	http.HandleFunc("GET /api/pipelines/{pipeline}/status", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) })
	http.HandleFunc("POST /api/pipelines/{pipeline}/run", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) })
	http.HandleFunc("PATCH /api/pipelines/{pipeline}/enable", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) })
	http.HandleFunc("GET /api/runtime/status", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) })
	http.HandleFunc("GET /api/runtime/{batch}/results/list", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) })
	http.HandleFunc("GET /api/runtime/{batch}/results/{result}/list", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) })
	http.HandleFunc("GET /api/runtime/{batch}/status", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) })
	http.HandleFunc("GET /api/collections/list", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) })
	http.HandleFunc("GET /api/collections/{collection}", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) })
	http.HandleFunc("DELETE /api/collections/{collection}", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) })
	http.HandleFunc("GET /api/collections/{collection}/data/{name}/list", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotImplemented) })
	http.ListenAndServe(fmt.Sprint(":", *port), http.DefaultServeMux)
}
