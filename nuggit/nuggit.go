// Binary nuggit provides a CLI tool for working with Nuggit servers and databases.
package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/wenooij/nuggit/nuggit/pipes"
	"github.com/wenooij/nuggit/nuggit/resources"
	"github.com/wenooij/nuggit/nuggit/results"
	"github.com/wenooij/nuggit/nuggit/rules"
)

func main() {
	app := &cli.App{
		Name:    "nuggit",
		Usage:   "Nuggit is a declarative tool for IR and web scraping",
		Version: "1.0.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "backend_addr",
				Aliases:     []string{"a", "addr", "backend"},
				Value:       "http://localhost:9402",
				DefaultText: ":9402",
			},
		},
		Commands: []*cli.Command{
			pipes.Cmd,
			resources.Cmd,
			results.Cmd,
			rules.Cmd,
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
