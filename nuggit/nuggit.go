// Binary nuggit provides a CLI tool for working with Nuggit servers and databases.
package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/wenooij/nuggit/nuggit/pipe"
	"github.com/wenooij/nuggit/nuggit/resource"
	"github.com/wenooij/nuggit/nuggit/rule"
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
			pipe.Cmd,
			resource.Cmd,
			rule.Cmd,
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
