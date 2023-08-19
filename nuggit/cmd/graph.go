//go:build !windows

// TODO(wes): Provide a graph_windows.go command for windows support.

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wenooij/nuggit/graphs"
	"github.com/wenooij/nuggit/graphviz"
)

var graphFlags struct {
	Graph  string
	Format string
}

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Print a graphviz a DOT or SVG",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := graphs.FromFile(graphFlags.Graph)
		if err != nil {
			return fmt.Errorf("failed to load Graph from file: %v", err)
		}
		gz := graphviz.Grapher{Graph: g}
		switch strings.ToUpper(graphFlags.Format) {
		case "DOT":
			data, err := gz.DOT()
			if err != nil {
				return fmt.Errorf("failed to render graph DOT: %v", err)
			}
			fmt.Println(string(data))
			return nil
		case "SVG":
			data, err := gz.SVG()
			if err != nil {
				return fmt.Errorf("failed to render graph DOT: %v", err)
			}
			fmt.Println(string(data))
			return nil
		default:
			return fmt.Errorf("-format is not supported: %q", graphFlags.Format)
		}
	},
}

func init() {
	fs := graphCmd.PersistentFlags()
	fs.StringVarP(&graphFlags.Graph, "graph", "g", "", "Graph to load from a local file")
	fs.StringVarP(&graphFlags.Format, "format", "o", "svg", `Format output (accepts "DOT" or "SVG")`)
	graphCmd.MarkFlagRequired("program")
	rootCmd.AddCommand(graphCmd)
}
