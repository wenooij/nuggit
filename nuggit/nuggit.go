// Binary nuggit provides a CLI tool for working with Nuggit servers and databases.
package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/wenooij/nuggit/status"
)

func main() {
	app := &cli.App{
		Name:    "nuggit",
		Usage:   "Nuggit is a declarative tool for IR and web scraping",
		Version: "1.0.0",
		Commands: []*cli.Command{
			{
				Name:  "get",
				Usage: "Gets a resource from the server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "id",
						Usage:    "UUID of the resource",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					return status.ErrUnimplemented
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
