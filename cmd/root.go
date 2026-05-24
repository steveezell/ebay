package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "card-hunt",
	Short: "Search eBay for sports cards on your want list",
	Long: `card-hunt manages a want list of sports cards and searches eBay for active
listings and sold comps — powered by ebay-pp-cli (printingpress.dev).

Prerequisites:
  1. Install ebay-pp-cli:  go install github.com/mvanhorn/printing-press-library/library/commerce/ebay/cmd/ebay-pp-cli@latest
  2. Authenticate:         ebay-pp-cli auth login --chrome

Quick start:
  card-hunt add "2011 Topps Update Mike Trout RC" --max 500
  card-hunt search
  card-hunt watch`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(compsCmd)
	rootCmd.AddCommand(watchCmd)
}
