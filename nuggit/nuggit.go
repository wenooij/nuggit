// Binary nuggit provides a CLI tool for working with Nuggit servers and databases.
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    "nuggit",
		Usage:   "Nuggit is a declarative tool for IR and web scraping",
		Version: "1.0.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "server_addr",
				Aliases:     []string{"a", "addr"},
				DefaultText: "http://localhost:9402",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "get",
				Aliases: []string{"g"},
				Usage:   "Gets resources from the server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "id",
						Usage: "UUID of the resource",
					},
				},
				Subcommands: []*cli.Command{{
					Name:    "pipe",
					Aliases: []string{"p"},
					Usage:   "Gets pipes from the server",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:    "name",
							Aliases: []string{"n"},
							Usage:   "Name of the pipe",
						},
						&cli.StringFlag{
							Name:    "digest",
							Aliases: []string{"d"},
							Usage:   "Digest of the pipe",
						},
						&cli.StringFlag{
							Name:    "name_digest",
							Aliases: []string{"nd"},
							Usage:   "Name@Digest of the pipe",
						},
					},
					Action: func(c *cli.Context) error {
						req, err := http.NewRequestWithContext(c.Context, "GET", c.String("server_addr")+"/api/pipes/"+c.String("name_digest"), nil)
						if err != nil {
							return err
						}
						resp, err := http.DefaultClient.Do(req)
						if err != nil {
							return err
						}
						body, err := io.ReadAll(resp.Body)
						defer resp.Body.Close()
						fmt.Println(string(body))
						return nil
					},
				}},
				Action: func(c *cli.Context) error {
					return fmt.Errorf("use a subcommand")
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
