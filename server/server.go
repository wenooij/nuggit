package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/status"
	"github.com/wenooij/nuggit/storage"
	_ "modernc.org/sqlite"
)

type server struct {
	*api.API
}

type serverSettings struct {
	port         int
	nuggitDir    string
	databasePath string
}

func NewServer(settings *serverSettings, r *gin.Engine, db *sql.DB) (*server, error) {
	// Check nuggit dir.
	info, err := os.Stat(settings.nuggitDir)
	if err != nil {
		return nil, fmt.Errorf("failed to access nuggit directory: %v: %w", err, status.ErrFailedPrecondition)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("nuggit path is not a directory (%v): %w", settings.nuggitDir, status.ErrFailedPrecondition)
	}

	collectionStore := storage.NewCollectionStore(db)
	pipeStore := storage.NewPipeStore(db)
	nodeStore := storage.NewNodeStore(db)

	api, err := api.NewAPI(collectionStore, pipeStore, nodeStore)
	if err != nil {
		return nil, err
	}
	s := &server{
		API: api,
	}
	s.registerAPI(r)
	return s, nil
}

func (s *server) registerAPI(r *gin.Engine) {
	var routes []string
	r.GET("/api/list", func(c *gin.Context) { c.JSON(http.StatusOK, routes) })
	r.GET("/api", func(c *gin.Context) { c.Redirect(http.StatusTemporaryRedirect, "list") })
	r.GET("/api/status", func(c *gin.Context) { status.WriteResponse(c, struct{}{}, nil) })
	s.registerActionsAPI(r)
	s.registerCollectionsAPI(r)
	s.registerNodesAPI(r)
	s.registerPipesAPI(r)
	s.registerResourcesAPI(r)
	s.registerRuntimeAPI(r)
	s.registerTriggerAPI(r)

	for _, r := range r.Routes() {
		routes = append(routes, fmt.Sprintf("%s %s", r.Method, r.Path))
	}
	slices.Sort(routes)
}

func (s *server) registerActionsAPI(r *gin.Engine) {
	r.GET("/api/actions/list", func(c *gin.Context) { status.WriteError(c, status.ErrUnimplemented) })
	r.GET("/api/actions/builtin_list", func(c *gin.Context) {
		resp, err := s.ListBuiltinActions(c.Request.Context(), &api.ListBuiltinActionsRequest{})
		status.WriteResponse(c, resp, err)
	})
	r.PUT("/api/actions/run", func(c *gin.Context) {
		req := new(api.RunActionRequest)
		if !status.ReadRequest(c, req) {
			return
		}
		resp, err := s.RunAction(c.Request.Context(), req)
		status.WriteResponse(c, resp, err)
	})
}

func (s *server) registerCollectionsAPI(r *gin.Engine) {
	r.GET("/api/collections/list", func(c *gin.Context) {
		resp, err := s.ListCollections(c.Request.Context(), &api.ListCollectionsRequest{})
		status.WriteResponse(c, resp, err)
	})
	r.POST("/api/collections", func(c *gin.Context) {
		req := new(api.CreateCollectionRequest)
		if !status.ReadRequest(c, req) {
			return
		}
		resp, err := s.CreateCollection(c.Request.Context(), req)
		status.WriteResponse(c, resp, err)
	})
	r.GET("/api/collections/:collection", func(c *gin.Context) {
		resp, err := s.GetCollection(c.Request.Context(), &api.GetCollectionRequest{Collection: c.Param("collection")})
		status.WriteResponse(c, resp, err)
	})
	r.DELETE("/api/collections/:collection", func(c *gin.Context) { status.WriteError(c, status.ErrUnimplemented) })
	r.GET("/api/collections/:collection/point/:name", func(c *gin.Context) { status.WriteError(c, status.ErrUnimplemented) })
	r.DELETE("/api/collections/:collection/point/:name", func(c *gin.Context) { status.WriteError(c, status.ErrUnimplemented) })
	r.GET("/api/collections/:collection/point/:name/list", func(c *gin.Context) { status.WriteError(c, status.ErrUnimplemented) })
}

func (s *server) registerNodesAPI(r *gin.Engine) {
	r.GET("/api/nodes/list", func(c *gin.Context) {
		resp, err := s.ListNodes(c.Request.Context(), &api.ListNodesRequest{})
		status.WriteResponse(c, resp, err)
	})
	r.GET("/api/nodes/:node", func(c *gin.Context) {
		resp, err := s.GetNode(c.Request.Context(), &api.GetNodeRequest{ID: c.Param("node")})
		status.WriteResponse(c, resp, err)
	})
	r.DELETE("/api/nodes/:node", func(c *gin.Context) {
		resp, err := s.DeleteNode(c.Request.Context(), &api.DeleteNodeRequest{ID: c.Param("node")})
		status.WriteResponse(c, resp, err)
	})
	r.POST("/api/nodes", func(c *gin.Context) {
		req := new(api.CreateNodeRequest)
		if !status.ReadRequest(c, req) {
			return
		}
		resp, err := s.CreateNode(c.Request.Context(), req)
		status.WriteResponse(c, resp, err)
	})
	r.GET("/api/nodes/orphans", func(c *gin.Context) {
		resp, err := s.ListOrphans(c.Request.Context(), &api.ListOrphansRequest{})
		status.WriteResponse(c, resp, err)
	})
	r.DELETE("/api/nodes/orphans", func(c *gin.Context) {
		resp, err := s.DeleteOrphans(c.Request.Context(), &api.DeleteOrphansRequest{})
		status.WriteResponse(c, resp, err)
	})
}

func (s *server) registerPipesAPI(r *gin.Engine) {
	r.GET("/api/pipes/list", func(c *gin.Context) {
		resp, err := s.ListPipes(c.Request.Context(), &api.ListPipesRequest{})
		status.WriteResponse(c, resp, err)
	})
	r.GET("/api/pipes/:pipe", func(c *gin.Context) {
		resp, err := s.GetPipe(c.Request.Context(), &api.GetPipeRequest{Pipe: c.Param("pipe")})
		status.WriteResponse(c, resp, err)
	})
	r.POST("/api/pipes", func(c *gin.Context) {
		req := new(api.CreatePipeRequest)
		if !status.ReadRequest(c, req) {
			return
		}
		resp, err := s.CreatePipe(c.Request.Context(), req)
		status.WriteResponse(c, resp, err)
	})
	r.GET("/api/pipes/:pipe/status", func(c *gin.Context) { status.WriteError(c, status.ErrUnimplemented) })
	r.PATCH("/api/pipes/:pipe/status", func(c *gin.Context) { status.WriteError(c, status.ErrUnimplemented) })
}

func (s *server) registerResourcesAPI(r *gin.Engine) {
	r.GET("/api/resources/list", func(c *gin.Context) { status.WriteError(c, status.ErrUnimplemented) })
	r.GET("/api/resources/versions/list", func(c *gin.Context) { status.WriteError(c, status.ErrUnimplemented) })
	r.PATCH("/api/resources", func(c *gin.Context) { status.WriteError(c, status.ErrUnimplemented) })
	r.POST("/api/resources", func(c *gin.Context) { status.WriteError(c, status.ErrUnimplemented) })
	r.PUT("/api/resources", func(c *gin.Context) { status.WriteError(c, status.ErrUnimplemented) })
	r.GET("/api/resources/:resource", func(c *gin.Context) { status.WriteError(c, status.ErrUnimplemented) })
	r.DELETE("/api/resources/:resource", func(c *gin.Context) { status.WriteError(c, status.ErrUnimplemented) })
}

func (s *server) registerRuntimeAPI(r *gin.Engine) {
	r.GET("/api/runtime/status", func(c *gin.Context) {
		resp, err := s.RuntimeStatus(c.Request.Context(), &api.RuntimeStatusRequest{})
		status.WriteResponse(c, resp, err)
	})
	r.GET("/api/runtimes/list", func(c *gin.Context) { status.WriteError(c, status.ErrUnimplemented) })
	r.GET("/api/runtimes/:runtime", func(c *gin.Context) { status.WriteError(c, status.ErrUnimplemented) })
	r.GET("/api/runtimes/:runtime/stats", func(c *gin.Context) { status.WriteError(c, status.ErrUnimplemented) })
	r.POST("/api/runtimes", func(c *gin.Context) { status.WriteError(c, status.ErrUnimplemented) })
}

func (s *server) registerTriggerAPI(r *gin.Engine) {
	r.POST("/api/trigger", func(c *gin.Context) {
		req := new(api.ImplicitTriggerRequest)
		if !status.ReadRequest(c, req) {
			return
		}
		resp, err := s.ImplicitTrigger(c.Request.Context(), req)
		status.WriteResponse(c, resp, err)
	})
	r.POST("/api/trigger/:collection", func(c *gin.Context) {
		resp, err := s.Trigger(c.Request.Context(), &api.TriggerRequest{Collection: c.Param("collection")})
		status.WriteResponse(c, resp, err)
	})
	r.POST("/api/trigger/batch", func(c *gin.Context) {
		req := new(api.TriggerBatchRequest)
		if !status.ReadRequest(c, req) {
			return
		}
		resp, err := s.TriggerBatch(c.Request.Context(), req)
		status.WriteResponse(c, resp, err)
	})
}

func main() {
	settings := &serverSettings{}
	flag.IntVar(&settings.port, "port", 9402, "Server port")
	flag.StringVar(&settings.nuggitDir, "nuggit_dir", filepath.Join(os.Getenv("HOME"), ".nuggit"), "Location of the Nuggit directory")
	flag.StringVar(&settings.databasePath, "database_path", filepath.Join(os.Getenv("HOME"), ".nuggit", "nuggit.sqlite"), "Sqllite database path")
	flag.Parse()

	db, err := sql.Open("sqlite", settings.databasePath)
	if err != nil {
		log.Printf("Failed to open sqlite database: %v", err)
		os.Exit(1)
	}
	if err := storage.InitDB(db); err != nil {
		log.Printf("Failed to initialized sqlite DB: %v", err)
		os.Exit(3)
	}
	db.SetMaxOpenConns(2) // https://pkg.go.dev/modernc.org/sqlite#section-readme

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))
	if _, err := NewServer(settings, r, db); err != nil {
		log.Printf("Initializing server failed: %v", err)
		os.Exit(4)
	}
	defer db.Close()
	r.Run(fmt.Sprint(":", settings.port))
}
