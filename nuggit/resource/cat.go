package resource

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/pipes"
	"github.com/wenooij/nuggit/resources"
	"gopkg.in/yaml.v3"
)

var catCmd = &cli.Command{
	Name:    "cat",
	Aliases: []string{"c"},
	Usage:   "Print resources received from files or stdin",
	Action: func(c *cli.Context) error {
		input := c.String("input")
		var in *os.File
		if input == "-" {
			in = os.Stdin
		} else {
			var err error
			in, err = os.Open(input)
			if err != nil {
				return err
			}
		}
		data, err := io.ReadAll(in)
		if err != nil {
			return err
		}

		var idx resources.Index
		if dir := c.String("dirs"); dir != "" {
			if err := idx.AddFS(os.DirFS(dir)); err != nil {
				return err
			}
		}

		tryFlatten := func(r *api.Resource) (*api.Resource, error) {
			if !c.Bool("flatten") {
				return r, nil
			}
			pipe := r.GetPipe()
			if pipe == nil {
				return r, nil
			}
			base, err := pipes.Flatten(idx.Pipes(), *pipe)
			if err != nil {
				return nil, err
			}
			r.Spec.(*api.Pipe).Pipe = base
			return r, nil
		}

		switch f := c.String("format"); f {
		case "json":
			// Validate, marshal indented.
			r := new(api.Resource)
			if err := json.Unmarshal(data, r); err != nil {
				return err
			}
			if r, err = tryFlatten(r); err != nil {
				return err
			}
			data, err := json.MarshalIndent(r, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		case "yaml":
			// Validate, convert to JSON.
			r := new(api.Resource)
			if err := yaml.Unmarshal(data, r); err != nil {
				return err
			}
			if r, err = tryFlatten(r); err != nil {
				return err
			}
			data, err := json.MarshalIndent(r, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		default:
			return fmt.Errorf("unknown format (%q)", f)
		}
	},
}
