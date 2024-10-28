package resource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/pipes"
	"github.com/wenooij/nuggit/resources"
)

var putCmd = &cli.Command{
	Name:    "put",
	Aliases: []string{"p"},
	Usage:   "Posts resources to the server",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "format",
			Aliases:     []string{"f"},
			DefaultText: "json",
			Value:       "json",
		},
		&cli.StringFlag{
			Name:        "input",
			Aliases:     []string{"in", "i"},
			DefaultText: "stdin",
			Value:       "-",
		},
		&cli.StringFlag{
			Name:    "dirs",
			Aliases: []string{"d"},
		},
		&cli.BoolFlag{
			Name:    "flatten",
			Aliases: []string{"t"},
			Value:   true,
		},
		&cli.BoolFlag{
			Name:    "replace",
			Aliases: []string{"r"},
			Usage:   "Replace other pipelines with the same name",
		},
	},
	Action: func(c *cli.Context) error {
		var idx resources.Index
		if dir := c.String("dirs"); dir != "" {
			if err := idx.AddFS(os.DirFS(dir)); err != nil {
				return err
			}
		}

		addr := c.String("server_addr")
		if addr == "" {
			return fmt.Errorf("-server_addr required")
		}

		u, err := url.JoinPath(addr, "/api/resources")
		if err != nil {
			return err
		}

		uniquePipes := idx.GetUniquePipes()

		for _, r := range idx.Entries {
			cr := api.CreateResourceRequest{}
			cr.Resource = r

			// Flatten pipes.
			if c.Bool("flatten") && r.Kind == api.KindPipe {
				flatPipe, err := pipes.Flatten(uniquePipes, *r.GetPipe())
				if err != nil {
					return err
				}
				r.ReplaceSpec(&flatPipe)
			}

			data, err := json.Marshal(cr)
			if err != nil {
				return err
			}
			req, err := http.NewRequest("POST", u, bytes.NewReader(data))
			if err != nil {
				return err
			}
			if _, err := http.DefaultClient.Do(req); err != nil {
				return err
			}
		}
		return nil
	},
}
