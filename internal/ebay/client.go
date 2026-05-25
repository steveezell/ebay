package ebay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Listing struct {
	ItemID    string  `json:"item_id"`
	Title     string  `json:"title"`
	Price     float64 `json:"price"`
	Currency  string  `json:"currency"`
	Condition string  `json:"condition"`
	Seller    string  `json:"seller"`
	Bids      int     `json:"bids"`
	BestOffer bool    `json:"best_offer"`
	Auction   bool    `json:"auction"`
	BIN       bool    `json:"buy_it_now"`
	TimeLeft  string  `json:"time_left"`
	URL       string  `json:"url"`
	Shipping  string  `json:"shipping"`
}

type CompStats struct {
	Query          string  `json:"query"`
	WindowDays     int     `json:"window_days"`
	SampleSize     int     `json:"sample_size"`
	UsedSize       int     `json:"used_size"`
	Mean           float64 `json:"mean"`
	Median         float64 `json:"median"`
	Min            float64 `json:"min"`
	Max            float64 `json:"max"`
	StdDev         float64 `json:"std_dev"`
	P25            float64 `json:"p25"`
	P75            float64 `json:"p75"`
	OutliersTrimmed int    `json:"outliers_trimmed"`
	Currency       string  `json:"currency"`
}

func conditionToCode(cond string) string {
	switch strings.ToLower(cond) {
	case "new":
		return "1000"
	case "used":
		return "3000"
	case "refurb":
		return "2000"
	case "parts":
		return "5000"
	default:
		return ""
	}
}

func MatchesCondition(listing Listing, filter string) bool {
	if filter == "" {
		return true
	}
	return strings.Contains(strings.ToLower(listing.Condition), strings.ToLower(filter))
}

func SearchListings(query string, maxPrice float64, condition string, binOnly bool) ([]Listing, error) {
	args := []string{"listings", "--json", "--agent", "--nkw", query}
	if maxPrice > 0 {
		args = append(args, "--udhi", strconv.FormatFloat(maxPrice, 'f', 2, 64))
	}
	if code := conditionToCode(condition); code != "" {
		args = append(args, "--lh-item-condition", code)
	}
	if binOnly {
		args = append(args, "--lh-bin", "1")
	}

	out, stderr, err := run("ebay-pp-cli", args...)
	if err != nil {
		return nil, wrapError("listings", stderr, err)
	}

	listings, parseErr := parseListings(out)
	if parseErr != nil {
		return nil, fmt.Errorf("parse listings: %w", parseErr)
	}

	if condition != "" {
		var filtered []Listing
		for _, l := range listings {
			if MatchesCondition(l, condition) {
				filtered = append(filtered, l)
			}
		}
		return filtered, nil
	}
	return listings, nil
}

func GetComps(query string) (*CompStats, error) {
	out, stderr, err := run("ebay-pp-cli", "comp", "--json", "--agent", query)
	if err != nil {
		return nil, wrapError("comp", stderr, err)
	}

	var stats CompStats
	if jsonErr := json.Unmarshal(out, &stats); jsonErr != nil {
		return nil, fmt.Errorf("parse comp response: %w", jsonErr)
	}
	return &stats, nil
}

func parseListings(data []byte) ([]Listing, error) {
	var envelope struct {
		Results []Listing `json:"results"`
	}
	if err := json.Unmarshal(data, &envelope); err == nil && envelope.Results != nil {
		return envelope.Results, nil
	}
	var items []Listing
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func run(name string, args ...string) (stdout []byte, stderr []byte, err error) {
	cmd := exec.Command(name, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return outBuf.Bytes(), errBuf.Bytes(), err
}

func wrapError(command string, stderr []byte, err error) error {
	msg := strings.TrimSpace(string(stderr))
	if msg == "" {
		return fmt.Errorf("ebay-pp-cli %s: %w", command, err)
	}
	first := strings.SplitN(msg, "\n", 2)[0]
	return fmt.Errorf("%s\n  (run 'ebay-pp-cli doctor' if auth is the issue)", first)
}
