package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	ebay "github.com/steveezell/ebay/internal/ebay"
	"github.com/steveezell/ebay/internal/wantlist"
)

var watchInterval time.Duration
var watchCard string

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Monitor eBay and alert on new listings",
	Long: `Poll eBay on a schedule and print any new listings that match your
want list criteria (card name, price cap, condition). Tracks seen listing IDs
so you only see each listing once.

Press Ctrl+C to stop watching.`,
	Example: `  card-hunt watch
  card-hunt watch --interval 10m
  card-hunt watch --card "Mike Trout" --interval 5m`,
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
		if watchCard != "" {
			cards = filterCards(wl.Cards, watchCard)
			if len(cards) == 0 {
				return fmt.Errorf("no card matching %q on your want list", watchCard)
			}
		}

		seen := loadSeen()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

		fmt.Printf("Watching %d card(s) — checking every %s. Press Ctrl+C to stop.\n\n", len(cards), watchInterval)

		runCheck := func() {
			ts := time.Now().Format("2006-01-02 15:04:05")
			newCount := 0
			for _, card := range cards {
				listings, err := ebay.SearchListings(card.Query, card.MaxPrice, card.Condition)
				if err != nil {
					fmt.Printf("[%s] Error checking %q: %v\n", ts, card.Name, err)
					continue
				}
				for _, l := range listings {
					key := card.Name + ":" + l.ItemID
					if seen[key] {
						continue
					}
					seen[key] = true
					newCount++
					printNewListing(ts, card.Name, card.MaxPrice, l)
				}
			}
			if newCount == 0 {
				fmt.Printf("[%s] No new listings.\n", ts)
			}
			saveSeen(seen)
		}

		// Run immediately on start.
		runCheck()

		ticker := time.NewTicker(watchInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				runCheck()
			case <-sig:
				fmt.Println("\nStopped.")
				return nil
			}
		}
	},
}

func printNewListing(ts, cardName string, maxPrice float64, l ebay.Listing) {
	ltype := listingType(l)
	bids := ""
	if l.Bids > 0 {
		bids = fmt.Sprintf(" | %d bids", l.Bids)
	}
	title := l.Title
	if len(title) > 70 {
		title = title[:67] + "..."
	}
	fmt.Printf("\n[%s] NEW LISTING — %s (max $%.2f)\n", ts, cardName, maxPrice)
	fmt.Printf("  $%.2f | %s | %s%s\n  %s\n  %s\n",
		l.Price, ltype, l.Condition, bids, title, l.URL)
}

func seenPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".card-hunt", "seen.json")
}

func loadSeen() map[string]bool {
	data, err := os.ReadFile(seenPath())
	if err != nil {
		return map[string]bool{}
	}
	var m map[string]bool
	if err := json.Unmarshal(data, &m); err != nil {
		return map[string]bool{}
	}
	return m
}

func saveSeen(m map[string]bool) {
	data, _ := json.Marshal(m)
	path := seenPath()
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	_ = os.WriteFile(path, data, 0644)
}

func init() {
	watchCmd.Flags().DurationVarP(&watchInterval, "interval", "i", 15*time.Minute, "How often to check eBay (e.g. 5m, 1h)")
	watchCmd.Flags().StringVarP(&watchCard, "card", "c", "", "Watch a specific card (partial name match)")
}
