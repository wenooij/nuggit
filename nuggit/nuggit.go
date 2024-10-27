// Binary nuggit provides a CLI tool for working with Nuggit servers and databases.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/api"
	"github.com/wenooij/nuggit/integrity"
	"github.com/wenooij/nuggit/pipes"
	"github.com/wenooij/nuggit/resources"
	"gopkg.in/yaml.v3"
)

func main() {
	app := &cli.App{
		Name:    "nuggit",
		Usage:   "Nuggit is a declarative tool for IR and web scraping",
		Version: "1.0.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "server_addr",
				Aliases:     []string{"a", "addr"},
				DefaultText: "http://localhost:9402",
			},
		},
		Commands: []*cli.Command{{
			Name:    "cat",
			Aliases: []string{"c"},
			Usage:   "Print resources received from files or stdin",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "format",
					Aliases: []string{"f"},
					Value:   "json",
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
			Subcommands: []*cli.Command{{
				Name:    "pipe",
				Aliases: []string{"p"},

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

					p := nuggit.Pipe{}

					tryFlatten := func() error {
						if !c.Bool("flatten") {
							return nil
						}
						var err error
						p, err = pipes.Flatten(idx.Pipes, p)
						if err != nil {
							return err
						}
						return err
					}

					switch f := c.String("format"); f {
					case "json":
						// Validate, marshal indented.
						if err := json.Unmarshal(data, &p); err != nil {
							return err
						}
						if err := tryFlatten(); err != nil {
							return err
						}
						data, err := json.MarshalIndent(p, "", "  ")
						if err != nil {
							return err
						}
						fmt.Println(string(data))
						return nil
					case "yaml":
						// Validate, convert to JSON.
						if err := yaml.Unmarshal(data, &p); err != nil {
							return err
						}
						if err := tryFlatten(); err != nil {
							return err
						}
						data, err := json.MarshalIndent(p, "", "  ")
						if err != nil {
							return err
						}
						fmt.Println(string(data))
						return nil
					default:
						return fmt.Errorf("unknown format (%q)", f)
					}
				},
			}, {
				Name:    "resource",
				Aliases: []string{"r"},
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
						base, err := pipes.Flatten(idx.GetUniquePipes(), *pipe)
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
			}, {
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
			}},
		}, {
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
			},
			Subcommands: []*cli.Command{{
				Name:    "resource",
				Aliases: []string{"r"},
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
			}, {
				Name:        "dummy",
				Description: "Print internal dummy text for a resource digest",
				Aliases:     []string{"p"},
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

					var sb strings.Builder
					if err := json.NewEncoder(&sb).Encode(r.GetSpec()); err != nil {
						return err
					}
					fmt.Println(sb.String())
					return nil
				},
			}},
		}, {
			Name:    "get",
			Aliases: []string{"g"},
			Usage:   "Gets resources from the server",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "id",
					Usage: "UUID of the resource",
				},
			},
			Subcommands: []*cli.Command{{
				Name:    "pipe",
				Aliases: []string{"p"},
				Usage:   "Gets pipes from the server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "name",
						Aliases: []string{"n"},
						Usage:   "Name of the pipe",
					},
					&cli.StringFlag{
						Name:    "digest",
						Aliases: []string{"d"},
						Usage:   "Digest of the pipe",
					},
					&cli.StringFlag{
						Name:    "name_digest",
						Aliases: []string{"nd"},
						Usage:   "Name@Digest of the pipe",
					},
				},
				Action: func(c *cli.Context) error {
					req, err := http.NewRequestWithContext(c.Context, "GET", c.String("server_addr")+"/api/pipes/"+c.String("name_digest"), nil)
					if err != nil {
						return err
					}
					resp, err := http.DefaultClient.Do(req)
					if err != nil {
						return err
					}
					body, err := io.ReadAll(resp.Body)
					defer resp.Body.Close()
					fmt.Println(string(body))
					return nil
				},
			}},
			Action: func(c *cli.Context) error {
				return fmt.Errorf("use a subcommand")
			},
		}},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
