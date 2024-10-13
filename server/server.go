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
	rt, err := runtime.NewRuntime()
	if err != nil {
		log.Printf("Initializing runtime failed: %v", err)
		os.Exit(3)
	}
	mux := http.NewServeMux()
	a := API{mux: mux, rt: rt}
	a.RegisterAPI()

	http.ListenAndServe(fmt.Sprint(":", *port), mux)
}
