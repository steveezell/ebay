package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/steveezell/ebay/internal/wantlist"
)

var removeCmd = &cobra.Command{
	Use:     "remove <card-name>",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a card from your want list",
	Long:    `Remove a card from your want list. Partial name matching is supported.`,
	Example: `  card-hunt remove "Mike Trout"
  card-hunt rm "Tatis"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		wl, err := wantlist.Load()
		if err != nil {
			return err
		}
		name, ok := wl.Remove(args[0])
		if !ok {
			return fmt.Errorf("no card matching %q found on your want list\n  Run 'card-hunt list' to see your cards", args[0])
		}
		fmt.Printf("Removed: %s\n", name)
		return nil
	},
}
