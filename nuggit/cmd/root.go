package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nuggit",
	Short: "Nuggit manages web scraper programs",
	Long: `Nuggit manages web scraper programs.
	Take a tour https://nuggit.dev/tour.
	Documentation is available at https://nuggit.dev/doc.`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func init() {
	rootCmd.AddCommand(
		sumCmd,
		versionCmd,
	)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
