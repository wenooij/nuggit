package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/status"
	"github.com/wenooij/nuggit/storage"
)

type server struct {
	*api.API
	*http.ServeMux
	patterns []string
}

type serverSettings struct {
	port      int
	nuggitDir string
	inMemory  bool
}

func NewServer(settings *serverSettings) (*server, error) {
	var store api.StoreInterface
	if settings.inMemory {
		store = storage.NewInMemory()
	} else {
		info, err := os.Stat(settings.nuggitDir)
		if err != nil {
			return nil, fmt.Errorf("failed to access nuggit directory: %v: %w", err, status.ErrFailedPrecondition)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("nuggit path is not a directory (%v): %w", settings.nuggitDir, status.ErrFailedPrecondition)
		}
		return nil, fmt.Errorf("persistent storage is not implemented; rerun with -in_memory: %w", status.ErrUnimplemented)
	}
	api, err := api.NewAPI(store)
	if err != nil {
		return nil, err
	}
	s := &server{
		API:      api,
		ServeMux: http.NewServeMux(),
	}
	s.registerAPI()
	return s, nil
}

func (s *server) handleFunc(pattern string, handler http.HandlerFunc) {
	s.patterns = append(s.patterns, pattern)
	s.ServeMux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
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

func (s *server) registerAPI() {
	s.handleFunc("GET /api/list", func(w http.ResponseWriter, r *http.Request) { status.WriteResponse(w, s.patterns, nil) })
	s.handleFunc("GET /api", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "list", http.StatusTemporaryRedirect)
	})
	s.handleFunc("GET /api/status", func(w http.ResponseWriter, r *http.Request) { status.WriteResponse(w, struct{}{}, nil) })
	s.registerCollectionsAPI()
	s.registerNodesAPI()
	s.registerPipesAPI()
	s.registerResourcesAPI()
	s.registerRuntimeAPI()
	slices.Sort(s.patterns)
}

func (s *server) registerActionsAPI() {
	s.handleFunc("GET /api/actions/builtin/list", func(w http.ResponseWriter, r *http.Request) {
		resp, err := s.ListBuiltinActions(&api.ListBuiltinActionsRequest{})
		status.WriteResponse(w, resp, err)
	})
}

func (s *server) registerCollectionsAPI() {
	s.handleFunc("GET /api/collections/list", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	s.handleFunc("GET /api/collections/{collection}", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	s.handleFunc("DELETE /api/collections/{collection}", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	s.handleFunc("GET /api/collections/{collection}/point/{name}", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	s.handleFunc("DELETE /api/collections/{collection}/point/{name}", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	s.handleFunc("GET /api/collections/{collection}/point/{name}/list", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
}

func (s *server) registerNodesAPI() {
	s.handleFunc("GET /api/nodes/list", func(w http.ResponseWriter, r *http.Request) {
		resp, err := s.ListNodes(&api.ListNodesRequest{})
		status.WriteResponse(w, resp, err)
	})
	s.handleFunc("GET /api/nodes/{node}", func(w http.ResponseWriter, r *http.Request) {
		resp, err := s.GetNode(&api.GetNodeRequest{ID: r.PathValue("node")})
		status.WriteResponse(w, resp, err)
	})
	s.handleFunc("GET /api/nodes/{node}/deps", func(w http.ResponseWriter, r *http.Request) {
		resp, err := s.GetNodeDependencies(&api.GetNodeDependenciesRequest{ID: r.PathValue("node")})
		status.WriteResponse(w, resp, err)
	})
	s.handleFunc("DELETE /api/nodes/{node}", func(w http.ResponseWriter, r *http.Request) {
		resp, err := s.DeleteNode(&api.DeleteNodeRequest{ID: r.PathValue("node")})
		status.WriteResponse(w, resp, err)
	})
	s.handleFunc("POST /api/nodes", func(w http.ResponseWriter, r *http.Request) {
		req := new(api.CreateNodeRequest)
		if !status.ReadRequest(w, r.Body, req) {
			return
		}
		resp, err := s.CreateNode(req)
		status.WriteResponse(w, resp, err)
	})
	s.handleFunc("GET /api/nodes/orphans", func(w http.ResponseWriter, r *http.Request) {
		resp, err := s.ListOrphans(&api.ListOrphansRequest{})
		status.WriteResponse(w, resp, err)
	})
	s.handleFunc("DELETE /api/nodes/orphans", func(w http.ResponseWriter, r *http.Request) {
		resp, err := s.DeleteOrphans(&api.DeleteOrphansRequest{})
		status.WriteResponse(w, resp, err)
	})
}

func (s *server) registerPipesAPI() {
	s.handleFunc("GET /api/pipes/list", func(w http.ResponseWriter, r *http.Request) {
		resp, err := s.ListPipes(&api.ListPipesRequest{})
		status.WriteResponse(w, resp, err)
	})
	s.handleFunc("GET /api/pipes/{pipe}", func(w http.ResponseWriter, r *http.Request) {
		resp, err := s.GetPipe(&api.GetPipeRequest{Pipe: r.PathValue("pipe")})
		status.WriteResponse(w, resp, err)
	})
	s.handleFunc("POST /api/pipes", func(w http.ResponseWriter, r *http.Request) {
		req := new(api.CreatePipeRequest)
		if !status.ReadRequest(w, r.Body, req) {
			return
		}
		resp, err := s.CreatePipe(req)
		status.WriteResponse(w, resp, err)
	})
	s.handleFunc("GET /api/pipes/{pipe}/status", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	s.handleFunc("PATCH /api/pipes/{pipe}/status", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
}

func (s *server) registerResourcesAPI() {
	s.handleFunc("GET /api/resources/list", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	s.handleFunc("GET /api/resources/versions/list", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	s.handleFunc("PATCH /api/resources", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	s.handleFunc("POST /api/resources", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	s.handleFunc("PUT /api/resources", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	s.handleFunc("GET /api/resources/{resource}", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	s.handleFunc("DELETE /api/resources/{resource}", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
}

func (s *server) registerRuntimeAPI() {
	s.handleFunc("GET /api/runtime/status", func(w http.ResponseWriter, r *http.Request) {
		resp, err := s.RuntimeStatus(&api.RuntimeStatusRequest{})
		status.WriteResponse(w, resp, err)
	})
	s.handleFunc("GET /api/runtimes/list", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	s.handleFunc("GET /api/runtimes/{runtime}", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	s.handleFunc("GET /api/runtimes/{runtime}/stats", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	s.handleFunc("POST /api/runtimes", func(w http.ResponseWriter, r *http.Request) { status.WriteError(w, status.ErrUnimplemented) })
	s.handleFunc("POST /api/runtimes/{runtime}/trigger", func(w http.ResponseWriter, r *http.Request) {
		req := new(api.ImplicitTriggerRequest)
		if !status.ReadRequest(w, r.Body, req) {
			return
		}
		resp, err := s.ImplicitTrigger(req)
		status.WriteResponse(w, resp, err)
	})
	s.handleFunc("POST /api/runtimes/{runtime}/trigger/{pipeline}", func(w http.ResponseWriter, r *http.Request) {
		req := new(api.TriggerRequest)
		if !status.ReadRequest(w, r.Body, req) {
			return
		}
		resp, err := s.Trigger(req)
		status.WriteResponse(w, resp, err)
	})
	s.handleFunc("POST /api/runtimes/{runtime}/trigger/batch", func(w http.ResponseWriter, r *http.Request) {
		req := new(api.TriggerBatchRequest)
		if !status.ReadRequest(w, r.Body, req) {
			return
		}
		resp, err := s.TriggerBatch(req)
		status.WriteResponse(w, resp, err)
	})
}

func main() {
	settings := &serverSettings{}
	flag.IntVar(&settings.port, "port", 9402, "Server port")
	flag.StringVar(&settings.nuggitDir, "nuggit_dir", filepath.Join(os.Getenv("HOME"), ".nuggit"), "Location of the Nuggit directory")
	flag.BoolVar(&settings.inMemory, "in_memory", false, "Whether to use in memory storage")
	flag.Parse()

	s, err := NewServer(settings)
	if err != nil {
		log.Printf("Initializing server failed: %v", err)
		os.Exit(3)
	}
	http.ListenAndServe(fmt.Sprint(":", settings.port), s)
}
