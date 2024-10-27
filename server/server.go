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
	"github.com/wenooij/nuggit/integrity"
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

	viewStore := storage.NewViewStore(db)
	pipeStore := storage.NewPipeStore(db)
	ruleStore := storage.NewRuleStore(db)
	planStore := storage.NewPlanStore(db)
	resultStore := storage.NewResultStore(db)
	resourceStore := storage.NewResourceStore(db)
	newTriggerPlanner := func() api.TriggerPlanner { return new(trigger.Planner) }

	api := api.NewAPI(viewStore, pipeStore, ruleStore, planStore, resultStore, resourceStore, newTriggerPlanner)
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
	s.registerViewsAPI(r)
	s.registerResourcesAPI(r)
	s.registerPipesAPI(r)
	s.registerTriggerAPI(r)

	for _, r := range r.Routes() {
		routes = append(routes, fmt.Sprintf("%s %s", r.Method, r.Path))
	}
	slices.Sort(routes)
}

func (s *server) registerViewsAPI(r *gin.Engine) {
	r.POST("/api/views", func(c *gin.Context) {
		req := new(api.CreateViewRequest)
		if !status.ReadRequest(c, req) {
			return
		}
		resp, err := s.CreateView(c.Request.Context(), req)
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
	r.POST("/api/triggers", func(c *gin.Context) {
		req := new(api.OpenTriggerRequest)
		if !status.ReadRequest(c, req) {
			return
		}
		resp, err := s.OpenTrigger(c.Request.Context(), req)
		if resp != nil && resp.Trigger != nil {
			status.WriteResponseStatusCode(c, http.StatusCreated, resp, err)
			return
		}
		status.WriteResponse(c, resp, err)
	})
	r.POST("/api/triggers/rules", func(c *gin.Context) {
		req := new(api.CreateRuleRequest)
		if !status.ReadRequest(c, req) {
			return
		}
		resp, err := s.CreateRule(c.Request.Context(), req)
		status.WriteResponse(c, resp, err)
	})
	r.POST("/api/triggers/exchange", func(c *gin.Context) {
		req := new(api.ExchangeResultsRequest)
		if !status.ReadRequest(c, req) {
			return
		}
		resp, err := s.ExchangeResults(c.Request.Context(), req)
		status.WriteResponse(c, resp, err)
	})
	r.POST("/api/triggers/close", func(c *gin.Context) {
		req := new(api.CloseTriggerRequest)
		if !status.ReadRequest(c, req) {
			return
		}
		resp, err := s.CloseTrigger(c.Request.Context(), req)
		status.WriteResponse(c, resp, err)
	})
}

func (s *server) registerResourcesAPI(r *gin.Engine) {
	r.POST("/api/resources", func(c *gin.Context) {
		req := new(api.CreateResourceRequest)
		if !status.ReadRequest(c, req) {
			return
		}
		resp, err := s.CreateResource(c.Request.Context(), req)
		status.WriteResponse(c, resp, err)
	})
}

func queryName(arg string) (integrity.NameDigest, error) {
	nameDigest, err := integrity.ParseNameDigest(arg)
	if err != nil {
		return nil, err
	}
	return nameDigest, nil
}

func queryNames(args []string) ([]integrity.NameDigest, error) {
	var names []integrity.NameDigest
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
