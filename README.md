# card-hunt

Search eBay for sports cards on your want list — powered by [printingpress.dev](https://printingpress.dev)'s `ebay-pp-cli`.

## What it does

- **Want list** — maintain a list of cards you're hunting with a max price cap per card
- **Search** — search eBay for active listings across your entire want list at once
- **Comps** — fetch 90-day sold-listing price stats to know what a card is actually worth
- **Watch** — poll eBay on a schedule and alert you the moment a new listing appears within your budget

## Prerequisites

**1. Install ebay-pp-cli** (requires Go 1.24+):
```sh
go install github.com/mvanhorn/printing-press-library/library/commerce/ebay/cmd/ebay-pp-cli@latest
```

**2. Authenticate with your eBay account** (reads cookies from Chrome):
```sh
ebay-pp-cli auth login --chrome
ebay-pp-cli doctor   # verify it worked
```

## Install card-hunt

```sh
go install github.com/steveezell/ebay/cmd/card-hunt@latest
```

Or build from source:
```sh
git clone https://github.com/steveezell/ebay
cd ebay
go build -o card-hunt ./cmd/card-hunt/
```

## Usage

### Build your want list

```sh
# Add a card with a price cap
card-hunt add "2011 Topps Update Mike Trout RC" --max 500

# Add with condition and notes
card-hunt add "2011 Topps Update Mike Trout RC" --max 300 \
  --condition used \
  --notes "PSA 9 or BGS 9.5 only"

# Use a custom search query (when card name alone isn't specific enough)
card-hunt add "Trout Update RC PSA 10" --max 800 \
  --query "2011 Topps Update Mike Trout RC PSA 10"

# View your list
card-hunt list

# Remove a card (partial name match works)
card-hunt remove "Mike Trout"
```

Condition options: `new`, `used`, `refurb`, `parts`

### Search eBay

```sh
# Search for all cards on your want list
card-hunt search

# Search for a single card
card-hunt search --card "Trout"
```

### Price comps (sold listings)

```sh
# 90-day sold comps for all cards
card-hunt comps

# Comps for one card
card-hunt comps --card "Luis Robert"
```

### Watch for new listings

```sh
# Check every 15 minutes (default)
card-hunt watch

# Check every 5 minutes, specific card
card-hunt watch --interval 5m --card "Mike Trout"
```

Press **Ctrl+C** to stop. Seen listing IDs are stored in `~/.card-hunt/seen.json` so you only see each listing once.

## Data storage

| File | Purpose |
|------|---------|
| `~/.card-hunt/wantlist.yaml` | Your want list |
| `~/.card-hunt/seen.json` | Listing IDs already shown by `watch` |
