package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wenooij/nuggit"
)

var sumFlags struct {
	Resource string
	Error    bool
	Quiet    bool
	Verbose  bool
	CRC32    string
	SHA1     string
	SHA2     string
}

func init() {
	fs := sumCmd.PersistentFlags()
	fs.StringVarP(&sumFlags.Resource, "resource", "r", "", "Resource to load from a local file.")
	fs.BoolVarP(&sumFlags.Error, "error", "e", false, "Similar to --quiet but output only the formatted error")
	fs.BoolVarP(&sumFlags.Quiet, "quiet", "q", false, "Output simply PASS or FAIL and nothing else.")
	fs.BoolVarP(&sumFlags.Verbose, "verbose", "v", false, "Output the hex encoded checksums.")
	fs.StringVar(&sumFlags.CRC32, "crc32", "", "expected CRC32 checksum.")
	fs.StringVar(&sumFlags.SHA1, "sha1", "", "expected SHA1 checksum.")
	fs.StringVar(&sumFlags.SHA2, "sha2", "", "expected SHA2 checksum.")
	sumCmd.MarkFlagRequired("resource")
}

var sumCmd = &cobra.Command{
	Use:   "sum",
	Short: "Checksum a Nuggit Resource",
	RunE:  runSumCmd,
}

func runSumCmd(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(sumFlags.Resource)
	if err != nil {
		return fmt.Errorf("failed to read Resource from file: %v", err)
	}
	sum := nuggit.Sums{
		CRC32: sumFlags.CRC32,
		SHA1:  sumFlags.SHA1,
		SHA2:  sumFlags.SHA2,
	}
	other := nuggit.Checksum(data)
	tests := sum.Test(other)

	if sumFlags.Quiet {
		if tests.Fail() {
			fmt.Printf("FAIL\n")
			return nil
		}
		fmt.Printf("PASS\n")
		return nil
	}

	if sumFlags.Error {
		if err, ok := tests.FormatError(); ok {
			fmt.Printf("%s\n", err)
		}
		return nil
	}

	fmt.Print(tests.Format(sumFlags.Verbose))
	return nil
}
