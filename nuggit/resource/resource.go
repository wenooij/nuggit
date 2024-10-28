package resource

import "github.com/urfave/cli/v2"

var Cmd = &cli.Command{
	Name:    "resource",
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
		&cli.BoolFlag{
			Name:    "flatten",
			Aliases: []string{"t"},
		},
	},
	Subcommands: []*cli.Command{
		catCmd,
		digestCmd,
		indexCmd,
		putCmd,
	},
}
