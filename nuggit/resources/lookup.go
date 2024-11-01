package resources

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/urfave/cli/v2"
)

var lookupCmd = &cli.Command{
	Name:    "lookup",
	Aliases: []string{"l"},
	Usage:   "Gets resources from the server",
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
			Name:    "uuid",
			Aliases: []string{"id", "u"},
			Usage:   "UUID of the resource",
		},
		&cli.StringFlag{
			Name:    "kind",
			Aliases: []string{"k"},
		},
	},
	Action: func(c *cli.Context) error {
		u, err := url.JoinPath(c.String("backend_addr"), "/api/resources/lookup")
		req, err := http.NewRequestWithContext(c.Context, "POST", u, nil)
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
}
