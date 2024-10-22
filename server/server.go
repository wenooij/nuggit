package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/status"
	"github.com/wenooij/nuggit/storage"
	"github.com/wenooij/nuggit/trigger"
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
	triggerStore := storage.NewTriggerStore(db)
	resultStore := storage.NewResultStore(db)
	newTriggerPlanner := func() api.TriggerPlanner { return new(trigger.Planner) }

	api := api.NewAPI(collectionStore, pipeStore, triggerStore, newTriggerPlanner, resultStore)
	s := &server{
		API: api,
	}
	s.registerAPI(r)
	return s, nil
}

func (s *server) registerAPI(r *gin.Engine) {
	var routes []string
	r.GET("/api/list", func(c *gin.Context) { c.JSON(http.StatusOK, routes) })
	r.GET("/api", func(c *gin.Context) { c.Redirect(http.StatusTemporaryRedirect, "/api/list") })
	r.GET("/api/status", func(c *gin.Context) { status.WriteResponse(c, struct{}{}, nil) })
	s.registerCollectionsAPI(r)
	s.registerPipesAPI(r)
	s.registerTriggerAPI(r)

	for _, r := range r.Routes() {
		routes = append(routes, fmt.Sprintf("%s %s", r.Method, r.Path))
	}
	slices.Sort(routes)
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
		nameDigest, err := api.ParseNameDigest(c.Param("collection"))
		if err != nil {
			status.WriteError(c, err)
			return
		}
		resp, err := s.GetCollection(c.Request.Context(), &api.GetCollectionRequest{Collection: &nameDigest})
		status.WriteResponse(c, resp, err)
	})
	r.DELETE("/api/collections/:collection", func(c *gin.Context) {
		nameDigest, err := api.ParseNameDigest(c.Param("collection"))
		if err != nil {
			status.WriteError(c, err)
			return
		}
		resp, err := s.DeleteCollection(c.Request.Context(), &api.DeleteCollectionRequest{Collection: &nameDigest})
		status.WriteResponse(c, resp, err)
	})
	r.DELETE("/api/collections", func(c *gin.Context) {
		req := new(api.DeleteCollectionsBatchRequest)
		if !status.ReadRequest(c, req) {
			return
		}
		resp, err := s.DeleteCollectionsBatch(c.Request.Context(), req)
		status.WriteResponse(c, resp, err)
	})
}

func (s *server) registerPipesAPI(r *gin.Engine) {
	r.GET("/api/pipes/list", func(c *gin.Context) {
		resp, err := s.ListPipes(c.Request.Context(), &api.ListPipesRequest{})
		status.WriteResponse(c, resp, err)
	})
	r.GET("/api/pipes", func(c *gin.Context) {
		pipes, err := queryNames(c.QueryArray("pipes"))
		if err != nil {
			status.WriteError(c, err)
			return
		}
		req := &api.GetPipesBatchRequest{Pipes: pipes}
		resp, err := s.GetPipesBatch(c.Request.Context(), req)
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
	r.POST("/api/pipes/batch", func(c *gin.Context) {
		req := new(api.CreatePipesBatchRequest)
		if !status.ReadRequest(c, req) {
			return
		}
		resp, err := s.CreatePipesBatch(c.Request.Context(), req)
		status.WriteResponse(c, resp, err)
	})
}

func (s *server) registerTriggerAPI(r *gin.Engine) {
	r.POST("/api/trigger", func(c *gin.Context) {
		req := new(api.CreateTriggerPlanRequest)
		if !status.ReadRequest(c, req) {
			return
		}
		resp, err := s.CreateTriggerPlan(c.Request.Context(), req)
		if resp != nil && resp.Trigger != nil {
			status.WriteResponseStatusCode(c, http.StatusCreated, resp, err)
			return
		}
		status.WriteResponse(c, resp, err)
	})
	r.POST("/api/trigger/:trigger/exchange", func(c *gin.Context) {
		var results []api.ExchangeResult
		if !status.ReadRequest(c, &results) {
			return
		}
		req := &api.ExchangeResultsRequest{
			Trigger: c.Param("trigger"),
			Results: results,
		}
		resp, err := s.ExchangeResults(c.Request.Context(), req)
		status.WriteResponse(c, resp, err)
	})
	r.POST("/api/trigger/:trigger/collection/:collection/commit", func(c *gin.Context) {
		req := new(api.CommitCollectionRequest)
		if !status.ReadRequest(c, req) {
			return
		}
		resp, err := s.CommitCollection(c.Request.Context(), req)
		status.WriteResponse(c, resp, err)
	})
	r.POST("/api/trigger/:trigger/commit", func(c *gin.Context) {
		req := new(api.CommitTriggerRequest)
		if !status.ReadRequest(c, req) {
			return
		}
		resp, err := s.CommitTrigger(c.Request.Context(), req)
		status.WriteResponse(c, resp, err)
	})
}

func queryName(arg string) (api.NameDigest, error) {
	nameDigest, err := api.ParseNameDigest(arg)
	if err != nil {
		return api.NameDigest{}, err
	}
	return nameDigest, nil
}

func queryNames(args []string) ([]api.NameDigest, error) {
	var names []api.NameDigest
	for _, arg := range args {
		for _, s := range strings.Split(arg, ",") {
			nameDigest, err := queryName(strings.TrimSpace(s))
			if err != nil {
				return nil, err
			}
			names = append(names, nameDigest)
		}
	}
	return names, nil
}

func main() {
	settings := &serverSettings{}
	flag.IntVar(&settings.port, "port", 9402, "Server port")
	flag.StringVar(&settings.nuggitDir, "nuggit_dir", filepath.Join(os.Getenv("HOME"), ".nuggit"), "Location of the Nuggit directory")
	flag.StringVar(&settings.databasePath, "database_path", filepath.Join(os.Getenv("HOME"), ".nuggit", "nuggit.sqlite"), "Sqllite database path")
	flag.Parse()

	ctx := context.Background()
	db, err := sql.Open("sqlite", settings.databasePath)
	if err != nil {
		log.Printf("Failed to open sqlite database: %v", err)
		os.Exit(1)
	}
	if err := storage.InitDB(ctx, db); err != nil {
		log.Printf("Failed to initialized sqlite DB: %v", err)
		os.Exit(3)
	}
	db.SetMaxOpenConns(1) // https://pkg.go.dev/modernc.org/sqlite#section-readme

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
