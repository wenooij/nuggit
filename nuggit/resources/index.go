package resources

import (
	"fmt"
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

		for _, key := range slices.SortedFunc(idx.Keys(), integrity.CompareNameDigest) {
			fmt.Println(key)
		}

		return nil
	},
}
