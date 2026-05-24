package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/steveezell/ebay/internal/wantlist"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "Show your want list",
	RunE: func(cmd *cobra.Command, args []string) error {
		wl, err := wantlist.Load()
		if err != nil {
			return err
		}
		if len(wl.Cards) == 0 {
			fmt.Println("Your want list is empty.")
			fmt.Println("  Add a card:  card-hunt add \"2011 Topps Update Mike Trout RC\" --max 500")
			return nil
		}
		fmt.Printf("Want list (%d cards) — saved to %s\n\n", len(wl.Cards), wantlist.DefaultPath())
		tw := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
		fmt.Fprintln(tw, "  NAME\tMAX PRICE\tCONDITION\tLISTING\tNOTES")
		fmt.Fprintln(tw, "  ────\t─────────\t─────────\t───────\t─────")
		for i, c := range wl.Cards {
			cond := c.Condition
			if cond == "" {
				cond = "any"
			}
			listing := "all"
			if c.BINOnly {
				listing = "BIN only"
			}
			fmt.Fprintf(tw, "  %d. %s\t$%.2f\t%s\t%s\t%s\n", i+1, c.Name, c.MaxPrice, cond, listing, c.Notes)
		}
		return tw.Flush()
	},
}
