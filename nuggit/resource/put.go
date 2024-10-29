package resource

import (
	"errors"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/client"
	"github.com/wenooij/nuggit/integrity"
	"github.com/wenooij/nuggit/pipes"
	"github.com/wenooij/nuggit/resources"
	"github.com/wenooij/nuggit/status"
)

var putCmd = &cli.Command{
	Name:    "put",
	Aliases: []string{"p"},
	Usage:   "Posts resources to the server",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "replace",
			Aliases: []string{"r"},
			Usage:   "Replace other resources with the same name",
		},
	},
	Action: func(c *cli.Context) error {
		var idx resources.Index
		if dir := c.String("dirs"); dir != "" {
			if err := idx.AddFS(os.DirFS(dir)); err != nil {
				return err
			}
		}

		cli := client.NewClient(c.String("backend_addr"))

		if c.Bool("replace") {
			for r := range idx.Values() {
				if r.Kind != api.KindPipe {
					continue
				}
				// We don't actually delete anything here.
				// Just disable existing pipes by name.
				if err := cli.DisablePipe(r.GetName(), ""); err != nil {
					return err
				}
			}
		}

		for r := range idx.Values() {
			// Flatten pipe.
			if r.Kind == api.KindPipe {
				flatPipe, err := pipes.Flatten(idx.Pipes(), *r.GetPipe())
				if err != nil {
					return err
				}
				r.ReplaceSpec(&flatPipe)
				integrity.SetDigest(r)
			}
			err := cli.CreateResource(r)
			if err == nil {
				continue
			}
			if !errors.Is(err, status.ErrAlreadyExists) {
				return err
			}
			// Reenable this already existing pipeline.
			if r.Kind == api.KindPipe {
				if err := cli.EnablePipe(r.GetName(), r.GetDigest()); err != nil {
					return err
				}
			}
		}
		return nil
	},
}
