package pipe

import (
	"github.com/urfave/cli/v2"
	"github.com/wenooij/nuggit/client"
)

var enableCmd = &cli.Command{
	Name:    "enable",
	Aliases: []string{"e"},
	Action: func(c *cli.Context) error {
		cli := client.NewClient()
		return cli.EnablePipe(c.String("name"), c.String("digest"))
	},
}
