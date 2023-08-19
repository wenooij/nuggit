package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Nuggit tool",
	Long:  "Print the version number of Nuggit tool",
	Run: func(*cobra.Command, []string) {
		fmt.Println("v1alpha")
	},
}
