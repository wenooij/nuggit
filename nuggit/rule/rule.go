package rule

import "github.com/urfave/cli/v2"

var Cmd = &cli.Command{
	Name: "rule",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "hostname",
			Aliases: []string{"host", "h"},
		},
		&cli.StringFlag{
			Name:    "url_pattern",
			Aliases: []string{"pattern", "u"},
		},
	},
	Subcommands: []*cli.Command{
		createCmd,
		deleteCmd,
	},
}
