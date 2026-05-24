package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/steveezell/ebay/internal/wantlist"
)

var (
	addMaxPrice  float64
	addCondition string
	addNotes     string
	addQuery     string
	addBINOnly   bool
)

var addCmd = &cobra.Command{
	Use:   "add <card-name>",
	Short: "Add a card to your want list",
	Long: `Add a card to your want list with a price cap.

The card name is used as the eBay search query by default. Use --query to
specify a different search term (e.g. if the card name is long or needs
specific keywords to find the right listing).`,
	Example: `  card-hunt add "2011 Topps Update Mike Trout RC" --max 500
  card-hunt add "Trout Rookie BGS 9.5" --max 300 --condition used --notes "BGS 9.5 only"
  card-hunt add "Fernando Tatis Jr" --max 75 --query "2020 Topps Chrome Tatis RC" --condition used`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		query := addQuery
		if query == "" {
			query = name
		}
		wl, err := wantlist.Load()
		if err != nil {
			return err
		}
		card := wantlist.Card{
			Name:      name,
			Query:     query,
			MaxPrice:  addMaxPrice,
			Condition: addCondition,
			BINOnly:   addBINOnly,
			Notes:     addNotes,
			Added:     time.Now(),
		}
		if err := wl.Add(card); err != nil {
			return err
		}
		cond := addCondition
		if cond == "" {
			cond = "any"
		}
		binStr := ""
		if addBINOnly {
			binStr = "  |  BIN only"
		}
		fmt.Printf("Added to want list: %s\n  Max price: $%.2f  |  Condition: %s%s\n", name, addMaxPrice, cond, binStr)
		if addNotes != "" {
			fmt.Printf("  Notes: %s\n", addNotes)
		}
		return nil
	},
}

func init() {
	addCmd.Flags().Float64VarP(&addMaxPrice, "max", "m", 0, "Maximum price to pay (required)")
	addCmd.Flags().StringVarP(&addCondition, "condition", "c", "", "Condition filter: new, used, refurb, parts (default: any)")
	addCmd.Flags().StringVarP(&addNotes, "notes", "n", "", "Personal notes (grading, variant, etc.)")
	addCmd.Flags().StringVarP(&addQuery, "query", "q", "", "eBay search query (defaults to card name)")
	addCmd.Flags().BoolVarP(&addBINOnly, "bin", "b", false, "Buy It Now listings only (skip auctions)")
	_ = addCmd.MarkFlagRequired("max")
}
