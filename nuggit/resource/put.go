package resource

import (
	"errors"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/wenooij/nuggit/client"
	"github.com/wenooij/nuggit/resources"
	"github.com/wenooij/nuggit/status"
)

var putCmd = &cli.Command{
	Name:    "put",
	Aliases: []string{"p"},
	Usage:   "Posts resources to the server",
	Action: func(c *cli.Context) error {
		idx := new(resources.Index)
		if dir := c.String("dirs"); dir != "" {
			if err := idx.AddFS(os.DirFS(dir)); err != nil {
				return err
			}
			// Qualify the index by setting the digest on any resources
			// to the digest of the uniquely named resource in idx.
			qualified, err := idx.Qualified()
			if err != nil {
				return err
			}
			idx = qualified
		}

		cli := client.NewClient(c.String("backend_addr"))

		for r := range idx.Values() {
			err := cli.CreateResource(r)
			if err == nil {
				continue
			}
			if !errors.Is(err, status.ErrAlreadyExists) {
				return err
			}
		}
		return nil
	},
}
