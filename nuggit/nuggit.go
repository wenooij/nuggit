// Binary nuggit provides a CLI tool for working with Nuggit servers and databases.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/wenooij/nuggit/api"
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

					switch f := c.String("format"); f {
					case "json":
						// Validate, marshal indented.
						p := new(api.Pipe)
						if err := json.Unmarshal(data, p); err != nil {
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
						p := new(api.Pipe)
						if err := yaml.Unmarshal(data, p); err != nil {
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
			}},
		}, {
			Name:    "digest",
			Aliases: []string{"d"},
			Usage:   "Digest resources received from files or stdin",
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
					Name:    "sha1",
					Aliases: []string{"s"},
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

					p := new(api.Pipe)
					switch f := c.String("format"); f {
					case "json":
						if err := json.Unmarshal(data, p); err != nil {
							return err
						}
					case "yaml":
						if err := yaml.Unmarshal(data, p); err != nil {
							return err
						}
					default:
						return fmt.Errorf("unknown format (%q)", f)
					}

					name := p.GetName()
					digest, err := api.PipeDigestSHA1(p)
					if err != nil {
						return err
					}
					if sha1 := c.String("sha1"); sha1 != "" { // Verify.
						if sha1 == digest {
							fmt.Printf("OK   %s\n", name)
						} else {
							fmt.Printf("FAIL %s  # %s\n", name, digest)
						}
					} else { // Print name@digest.
						pipeDigest, err := api.JoinPipeDigest(name, digest)
						if err != nil {
							return err
						}
						fmt.Println(pipeDigest)
					}
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
