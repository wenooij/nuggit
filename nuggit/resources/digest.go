package resources

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/integrity"
	"gopkg.in/yaml.v3"
)

var digestCmd = &cli.Command{
	Name:    "digest",
	Aliases: []string{"d"},
	Usage:   "Digest resources received from files or stdin",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "format",
			Aliases:     []string{"f"},
			DefaultText: "json",
			Value:       "json",
		},
		&cli.StringFlag{
			Name:        "input",
			Aliases:     []string{"in", "i"},
			DefaultText: "stdin",
			Value:       "-",
		},
		&cli.StringFlag{
			Name:    "sha1",
			Aliases: []string{"s"},
		},
		&cli.BoolFlag{
			Name:    "dummy",
			Aliases: []string{"d"},
		},
	},
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

		r := new(api.Resource)

		switch f := c.String("format"); f {
		case "json":
			if err := json.Unmarshal(data, r); err != nil {
				return err
			}
		case "yaml":
			if err := yaml.Unmarshal(data, r); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown format (%q)", f)
		}

		digest, err := integrity.GetDigest(r)
		if err != nil {
			return err
		}

		if sha1 := c.String("sha1"); sha1 != "" { // Verify.
			if sha1 == digest {
				fmt.Printf("OK   %s\n", r.GetName())
			} else {
				fmt.Printf("FAIL %s  # %s\n", r.GetName(), digest)
			}
		} else { // Print name@digest.
			fmt.Println(integrity.FormatString(r))
		}
		return nil
	},
}
