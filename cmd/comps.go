package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	ebay "github.com/steveezell/ebay/internal/ebay"
	"github.com/steveezell/ebay/internal/wantlist"
)

var compsCard string

var compsCmd = &cobra.Command{
	Use:     "comps",
	Aliases: []string{"comp", "prices"},
	Short:   "Show sold-listing price comps for your want list",
	Long: `Fetch sold-listing price statistics from the last 90 days for every card
on your want list. Helps you know what a card is actually worth before buying.

Use --card to look up comps for a single card by name.`,
	Example: `  card-hunt comps
  card-hunt comps --card "Mike Trout"`,
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
		if compsCard != "" {
			cards = filterCards(wl.Cards, compsCard)
			if len(cards) == 0 {
				return fmt.Errorf("no card matching %q on your want list", compsCard)
			}
		}

		fmt.Printf("Fetching 90-day sold comps for %d card(s)...\n\n", len(cards))

		for i, card := range cards {
			fmt.Printf("[%d/%d] %s\n", i+1, len(cards), card.Name)
			if card.Notes != "" {
				fmt.Printf("       Note: %s\n", card.Notes)
			}

			stats, err := ebay.GetComps(card.Query)
			if err != nil {
				fmt.Printf("       Error: %v\n\n", err)
				continue
			}

			if stats.SampleSize == 0 {
				fmt.Println("       No sold listings found in the last 90 days.")
				fmt.Println()
				continue
			}

			printCompStats(stats, card.MaxPrice)
			fmt.Println()
		}
		return nil
	},
}

func printCompStats(s *ebay.CompStats, maxPrice float64) {
	verdict := ""
	if maxPrice > 0 {
		if s.Median <= maxPrice {
			verdict = "  ✓ median is within your budget"
		} else {
			pct := ((s.Median - maxPrice) / maxPrice) * 100
			verdict = fmt.Sprintf("  ✗ median is %.0f%% over your $%.2f cap", pct, maxPrice)
		}
	}

	fmt.Printf("       Sold comps (%d sales", s.SampleSize)
	if s.OutliersTrimmed > 0 {
		fmt.Printf(", %d outliers trimmed", s.OutliersTrimmed)
	}
	fmt.Printf(")%s\n", verdict)

	line1 := fmt.Sprintf("       Mean: $%-8.2f  Median: $%-8.2f  Range: $%.2f – $%.2f",
		s.Mean, s.Median, s.Min, s.Max)
	line2 := fmt.Sprintf("       P25:  $%-8.2f  P75:    $%.2f",
		s.P25, s.P75)
	fmt.Println(line1)
	fmt.Println(line2)

	if maxPrice > 0 && s.Max > 0 {
		fmt.Println()
		printPriceBar(s.Min, s.P25, s.Median, s.P75, s.Max, maxPrice)
	}
}

func printPriceBar(min, p25, median, p75, max, cap float64) {
	width := 40
	mark := func(val float64) int {
		if max == min {
			return 0
		}
		pos := int(((val - min) / (max - min)) * float64(width))
		if pos < 0 {
			pos = 0
		}
		if pos > width {
			pos = width
		}
		return pos
	}

	bar := []rune(strings.Repeat("─", width+1))
	for i := mark(p25); i <= mark(p75) && i <= width; i++ {
		bar[i] = '▓'
	}
	mPos := mark(median)
	if mPos >= 0 && mPos <= width {
		bar[mPos] = '◆'
	}
	capPos := mark(cap)
	label := "cap"
	if cap < min {
		label = "cap (below market)"
		capPos = 0
	} else if cap > max {
		label = "cap (above market)"
		capPos = width
	}
	if capPos >= 0 && capPos <= width {
		bar[capPos] = '|'
	}

	fmt.Printf("       $%.0f %s $%.0f\n", min, string(bar), max)
	fmt.Printf("              %*s %s ($%.2f)\n", capPos+len(fmt.Sprintf("$%.0f ", min)), "", label, cap)
}

func init() {
	compsCmd.Flags().StringVarP(&compsCard, "card", "c", "", "Look up comps for a specific card (partial name match)")
}
