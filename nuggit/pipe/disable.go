package pipe

import (
	"github.com/urfave/cli/v2"
	"github.com/wenooij/nuggit/client"
)

var disableCmd = &cli.Command{
	Name:    "disable",
	Aliases: []string{"d"},
	Action: func(c *cli.Context) error {
		cli := client.NewClient()
		return cli.DisablePipe(c.String("name"), c.String("digest"))
	},
}
