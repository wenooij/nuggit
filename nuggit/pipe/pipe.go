package pipe

import (
	"github.com/urfave/cli/v2"
)

var Cmd = &cli.Command{
	Name:    "pipe",
	Aliases: []string{"p"},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "name",
			Aliases: []string{"n"},
		},
		&cli.StringFlag{
			Name:    "digest",
			Aliases: []string{"d"},
		},
		&cli.StringSliceFlag{
			Name:    "labels",
			Aliases: []string{"l"},
		},
	},
	Subcommands: []*cli.Command{},
}
