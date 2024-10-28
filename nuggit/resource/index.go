package resource

import (
	"fmt"
	"maps"
	"os"
	"slices"

	"github.com/urfave/cli/v2"
	"github.com/wenooij/nuggit/integrity"
	"github.com/wenooij/nuggit/resources"
)

var indexCmd = &cli.Command{
	Name:    "index",
	Aliases: []string{"i"},
	Action: func(c *cli.Context) error {
		var idx resources.Index

		dirs := c.String("dirs")
		if dirs == "" {
			return nil
		}

		if err := idx.AddFS(os.DirFS(dirs)); err != nil {
			return err
		}

		for _, nd := range slices.SortedFunc(maps.Keys(idx.Entries), integrity.CompareNameDigest) {
			key := integrity.Key(nd)
			fmt.Println(key)
		}

		return nil
	},
}
