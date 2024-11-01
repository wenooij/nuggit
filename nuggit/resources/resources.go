package resources

import "github.com/urfave/cli/v2"

var Cmd = &cli.Command{
	Name:    "resources",
	Aliases: []string{"r"},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "format",
			Aliases:     []string{"f"},
			Value:       "json",
			DefaultText: "json",
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
	},
	Subcommands: []*cli.Command{
		catCmd,
		digestCmd,
		indexCmd,
		putCmd,
	},
}
