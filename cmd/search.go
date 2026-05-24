package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	ebay "github.com/steveezell/ebay/internal/ebay"
	"github.com/steveezell/ebay/internal/wantlist"
)

var searchCard string

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search eBay for cards on your want list",
	Long: `Search eBay for active listings matching every card on your want list.
Results are filtered to your price cap and condition preference.

Use --card to search for a single card by name (partial match supported).`,
	Example: `  card-hunt search
  card-hunt search --card "Mike Trout"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		wl, err := wantlist.Load()
		if err != nil {
			return err
		}
		if len(wl.Cards) == 0 {
			fmt.Println("Your want list is empty. Run 'card-hunt add' first.")
			return nil
		}

		cards := wl.Cards
		if searchCard != "" {
			cards = filterCards(wl.Cards, searchCard)
			if len(cards) == 0 {
				return fmt.Errorf("no card matching %q on your want list", searchCard)
			}
		}

		fmt.Printf("Searching eBay for %d card(s)...\n\n", len(cards))

		totalFound := 0
		for i, card := range cards {
			cond := card.Condition
			if cond == "" {
				cond = "any"
			}
			fmt.Printf("[%d/%d] %s\n      max $%.2f  |  condition: %s\n",
				i+1, len(cards), card.Name, card.MaxPrice, cond)

			listings, err := ebay.SearchListings(card.Query, card.MaxPrice, card.Condition, card.BINOnly)
			if err != nil {
				fmt.Printf("      Error: %v\n\n", err)
				continue
			}

			if len(listings) == 0 {
				fmt.Println("      No listings found within your criteria.")
				fmt.Println()
				continue
			}

			totalFound += len(listings)
			fmt.Printf("      %d listing(s) found:\n\n", len(listings))
			printListings(listings)
			fmt.Println()
		}

		if len(cards) > 1 {
			fmt.Printf("Total: %d listings across %d cards\n", totalFound, len(cards))
		}
		return nil
	},
}

func filterCards(cards []wantlist.Card, substr string) []wantlist.Card {
	lower := strings.ToLower(substr)
	var out []wantlist.Card
	for _, c := range cards {
		if strings.Contains(strings.ToLower(c.Name), lower) {
			out = append(out, c)
		}
	}
	return out
}

func printListings(listings []ebay.Listing) {
	tw := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "      PRICE\tTYPE\tCONDITION\tBIDS\tTITLE\tURL")
	for _, l := range listings {
		listingType := listingType(l)
		bids := "-"
		if l.Bids > 0 {
			bids = fmt.Sprintf("%d", l.Bids)
		}
		title := l.Title
		if len(title) > 55 {
			title = title[:52] + "..."
		}
		fmt.Fprintf(tw, "      $%.2f\t%s\t%s\t%s\t%s\t%s\n",
			l.Price, listingType, l.Condition, bids, title, l.URL)
	}
	_ = tw.Flush()
}

func listingType(l ebay.Listing) string {
	switch {
	case l.Auction && l.BIN:
		return "BIN+Auction"
	case l.Auction:
		return "Auction"
	case l.BIN:
		return "BIN"
	default:
		return "Listing"
	}
}

func init() {
	searchCmd.Flags().StringVarP(&searchCard, "card", "c", "", "Search for a specific card (partial name match)")
}
