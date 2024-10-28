package rule

import "github.com/urfave/cli/v2"

var createCmd = &cli.Command{
	Name:    "create",
	Aliases: []string{"c"},
	Action:  func(c *cli.Context) error { return nil },
}
