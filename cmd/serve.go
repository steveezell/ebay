package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/steveezell/ebay/internal/wantlist"
)

var servePort int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start a local web UI for managing your want list",
	Long:  `Starts a local HTTP server and opens a browser-based UI for managing your card want list.`,
	Example: `  card-hunt serve
  card-hunt serve --port 9090`,
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := fmt.Sprintf(":%d", servePort)

		mux := http.NewServeMux()
		mux.HandleFunc("/", handleUI)
		mux.HandleFunc("/api/cards", handleCards)
		mux.HandleFunc("/api/cards/", handleCardByName)

		fmt.Printf("card-hunt UI running at http://localhost:%d\nPress Ctrl+C to stop.\n", servePort)
		return http.ListenAndServe(addr, mux)
	},
}

func handleUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, uiHTML)
}

func handleCards(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		wl, err := wantlist.Load()
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cards := wl.Cards
		if cards == nil {
			cards = []wantlist.Card{}
		}
		jsonOK(w, cards)

	case http.MethodPost:
		var input struct {
			Name      string  `json:"name"`
			Query     string  `json:"query"`
			MaxPrice  float64 `json:"max_price"`
			Condition string  `json:"condition"`
			BINOnly   bool    `json:"bin_only"`
			Notes     string  `json:"notes"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			jsonError(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		input.Name = strings.TrimSpace(input.Name)
		if input.Name == "" {
			jsonError(w, "name is required", http.StatusBadRequest)
			return
		}
		if input.MaxPrice <= 0 {
			jsonError(w, "max_price must be greater than 0", http.StatusBadRequest)
			return
		}
		query := strings.TrimSpace(input.Query)
		if query == "" {
			query = input.Name
		}
		wl, err := wantlist.Load()
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		card := wantlist.Card{
			Name:      input.Name,
			Query:     query,
			MaxPrice:  input.MaxPrice,
			Condition: input.Condition,
			BINOnly:   input.BINOnly,
			Notes:     input.Notes,
			Added:     time.Now(),
		}
		if err := wl.Add(card); err != nil {
			jsonError(w, err.Error(), http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusCreated)
		jsonOK(w, card)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleCardByName(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/cards/")
	name = strings.TrimSpace(name)
	if name == "" {
		http.NotFound(w, r)
		return
	}

	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	wl, err := wantlist.Load()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	removed, ok := wl.Remove(name)
	if !ok {
		jsonError(w, fmt.Sprintf("card not found: %s", name), http.StatusNotFound)
		return
	}
	jsonOK(w, map[string]string{"removed": removed})
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func init() {
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8080, "Port to listen on")
	rootCmd.AddCommand(serveCmd)
}

const uiHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>card-hunt — Want List</title>
<style>
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    background: #0f1117;
    color: #e2e8f0;
    min-height: 100vh;
    padding: 2rem 1rem;
  }

  .container { max-width: 900px; margin: 0 auto; }

  header {
    display: flex;
    align-items: baseline;
    gap: 1rem;
    margin-bottom: 2rem;
    border-bottom: 1px solid #2d3748;
    padding-bottom: 1rem;
  }
  header h1 { font-size: 1.5rem; font-weight: 700; color: #f6e05e; letter-spacing: -0.5px; }
  header p  { color: #718096; font-size: 0.9rem; }

  /* Add form */
  .add-form {
    background: #1a1f2e;
    border: 1px solid #2d3748;
    border-radius: 10px;
    padding: 1.25rem 1.5rem;
    margin-bottom: 2rem;
  }
  .add-form h2 { font-size: 0.85rem; text-transform: uppercase; letter-spacing: 1px; color: #718096; margin-bottom: 1rem; }

  .form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; margin-bottom: 0.75rem; }
  .form-row.three { grid-template-columns: 1fr 1fr 1fr; }
  .form-full { margin-bottom: 0.75rem; }

  label { display: block; font-size: 0.75rem; color: #a0aec0; margin-bottom: 0.3rem; }
  input, select {
    width: 100%;
    background: #0f1117;
    border: 1px solid #2d3748;
    border-radius: 6px;
    color: #e2e8f0;
    padding: 0.5rem 0.75rem;
    font-size: 0.9rem;
    outline: none;
    transition: border-color 0.15s;
  }
  input:focus, select:focus { border-color: #f6e05e; }
  input::placeholder { color: #4a5568; }
  select option { background: #1a1f2e; }

  .form-actions { display: flex; justify-content: flex-end; margin-top: 0.25rem; }
  .btn {
    padding: 0.5rem 1.25rem;
    border-radius: 6px;
    font-size: 0.875rem;
    font-weight: 600;
    cursor: pointer;
    border: none;
    transition: opacity 0.15s;
  }
  .btn:hover { opacity: 0.85; }
  .btn-primary { background: #f6e05e; color: #1a1f2e; }
  .btn-danger  { background: transparent; color: #fc8181; border: 1px solid #fc8181; padding: 0.25rem 0.6rem; font-size: 0.8rem; }
  .btn-danger:hover { background: #fc818122; }

  /* Error banner */
  .error-banner {
    background: #2d1515;
    border: 1px solid #e53e3e;
    color: #fc8181;
    border-radius: 6px;
    padding: 0.6rem 1rem;
    margin-bottom: 1rem;
    font-size: 0.875rem;
    display: none;
  }

  /* Want list table */
  .list-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 0.75rem;
  }
  .list-header h2 { font-size: 1rem; font-weight: 600; }
  .badge {
    background: #2d3748;
    color: #a0aec0;
    border-radius: 999px;
    padding: 0.15rem 0.6rem;
    font-size: 0.75rem;
    font-weight: 600;
  }

  table { width: 100%; border-collapse: collapse; }
  .bin-badge {
    display: inline-block;
    font-size: 0.7rem;
    padding: 0.1rem 0.4rem;
    border-radius: 4px;
    background: #2a3d5c;
    color: #63b3ed;
    margin-left: 0.4rem;
    vertical-align: middle;
  }

  thead th {
    text-align: left;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: #718096;
    padding: 0.5rem 0.75rem;
    border-bottom: 1px solid #2d3748;
  }
  tbody tr {
    border-bottom: 1px solid #1a2035;
    transition: background 0.1s;
  }
  tbody tr:hover { background: #1a1f2e; }
  tbody td { padding: 0.75rem; font-size: 0.9rem; vertical-align: middle; }

  .card-name { font-weight: 500; color: #e2e8f0; }
  .card-notes { font-size: 0.78rem; color: #718096; margin-top: 0.2rem; }
  .price { font-family: monospace; color: #68d391; font-weight: 600; }
  .condition {
    display: inline-block;
    font-size: 0.75rem;
    padding: 0.15rem 0.5rem;
    border-radius: 4px;
    background: #2d3748;
    color: #a0aec0;
  }
  .condition.used   { background: #2c4a3e; color: #68d391; }
  .condition.new    { background: #2a3d5c; color: #63b3ed; }
  .condition.refurb { background: #3d2c4a; color: #b794f4; }
  .condition.any    { background: #2d3748; color: #718096; }

  .empty-state {
    text-align: center;
    padding: 3rem;
    color: #4a5568;
  }
  .empty-state p { margin-bottom: 0.5rem; }

  .spinner {
    text-align: center;
    padding: 2rem;
    color: #4a5568;
    font-size: 0.875rem;
  }
</style>
</head>
<body>
<div class="container">

  <header>
    <h1>card-hunt</h1>
    <p>eBay sports card want list</p>
  </header>

  <div class="add-form">
    <h2>Add a card</h2>
    <div id="errorBanner" class="error-banner"></div>
    <div class="form-full">
      <label for="name">Card name *</label>
      <input id="name" type="text" placeholder="1997 Ken Griffey Jr New Pinnacle Artist Proof" autocomplete="off">
    </div>
    <div class="form-row">
      <div>
        <label for="maxPrice">Max price ($) *</label>
        <input id="maxPrice" type="number" min="0.01" step="0.01" placeholder="150.00">
      </div>
      <div>
        <label for="condition">Condition</label>
        <select id="condition">
          <option value="">Any</option>
          <option value="used">Used</option>
          <option value="new">New</option>
          <option value="refurb">Refurb</option>
          <option value="parts">Parts / not working</option>
        </select>
      </div>
    </div>
    <div class="form-row three">
      <div>
        <label for="query">Search query <span style="color:#4a5568">(optional — defaults to card name)</span></label>
        <input id="query" type="text" placeholder="griffey new pinnacle artist proof 1997">
      </div>
      <div>
        <label for="notes">Notes</label>
        <input id="notes" type="text" placeholder="AP /100, not mirror gold">
      </div>
      <div style="display:flex;align-items:flex-end;padding-bottom:2px">
        <label style="display:flex;align-items:center;gap:0.5rem;cursor:pointer;margin:0">
          <input id="binOnly" type="checkbox" style="width:auto;accent-color:#f6e05e">
          <span style="font-size:0.875rem;color:#e2e8f0">Buy It Now only</span>
        </label>
      </div>
    </div>
    <div class="form-actions">
      <button class="btn btn-primary" onclick="addCard()">Add to want list</button>
    </div>
  </div>

  <div class="list-header">
    <h2>Want list <span class="badge" id="cardCount">0</span></h2>
  </div>

  <div id="listContainer">
    <div class="spinner">Loading…</div>
  </div>

</div>

<script>
async function loadCards() {
  try {
    const res = await fetch('/api/cards');
    const cards = await res.json();
    renderCards(cards);
  } catch (e) {
    document.getElementById('listContainer').innerHTML =
      '<div class="empty-state"><p>Could not reach the server.</p></div>';
  }
}

function renderCards(cards) {
  document.getElementById('cardCount').textContent = cards.length;
  const el = document.getElementById('listContainer');
  if (!cards.length) {
    el.innerHTML = '<div class="empty-state"><p>Your want list is empty.</p><p>Add your first card above.</p></div>';
    return;
  }
  const rows = cards.map(c => {
    const cond = c.condition || 'any';
    const condLabel = cond.charAt(0).toUpperCase() + cond.slice(1);
    const notes = c.notes ? '<div class="card-notes">' + esc(c.notes) + '</div>' : '';
    const queryNote = c.query && c.query !== c.name
      ? '<div class="card-notes">search: ' + esc(c.query) + '</div>' : '';
    const binBadge = c.bin_only ? '<span class="bin-badge">BIN</span>' : '';
    return '<tr>' +
      '<td><div class="card-name">' + esc(c.name) + binBadge + '</div>' + notes + queryNote + '</td>' +
      '<td><span class="price">$' + parseFloat(c.max_price).toFixed(2) + '</span></td>' +
      '<td><span class="condition ' + esc(cond) + '">' + esc(condLabel) + '</span></td>' +
      '<td><button class="btn btn-danger" data-card="' + esc(c.name) + '">Remove</button></td>' +
      '</tr>';
  }).join('');
  el.innerHTML = '<table>' +
    '<thead><tr><th>Card</th><th>Max Price</th><th>Condition</th><th></th></tr></thead>' +
    '<tbody>' + rows + '</tbody></table>';
}

async function addCard() {
  const name     = document.getElementById('name').value.trim();
  const maxPrice = parseFloat(document.getElementById('maxPrice').value);
  const condition = document.getElementById('condition').value;
  const query    = document.getElementById('query').value.trim();
  const notes    = document.getElementById('notes').value.trim();
  const binOnly  = document.getElementById('binOnly').checked;

  if (!name) { showError('Card name is required.'); return; }
  if (!maxPrice || maxPrice <= 0) { showError('Enter a valid max price.'); return; }

  hideError();
  try {
    const res = await fetch('/api/cards', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, max_price: maxPrice, condition, query, notes, bin_only: binOnly })
    });
    const data = await res.json();
    if (!res.ok) { showError(data.error || 'Something went wrong.'); return; }
    document.getElementById('name').value = '';
    document.getElementById('maxPrice').value = '';
    document.getElementById('condition').value = '';
    document.getElementById('query').value = '';
    document.getElementById('notes').value = '';
    document.getElementById('binOnly').checked = false;
    await loadCards();
  } catch (e) {
    showError('Could not reach the server.');
  }
}

async function removeCard(name) {
  if (!confirm('Remove "' + name + '" from your want list?')) return;
  try {
    const res = await fetch('/api/cards/' + encodeURIComponent(name), { method: 'DELETE' });
    if (!res.ok) {
      const data = await res.json();
      alert(data.error || 'Could not remove card.');
      return;
    }
    await loadCards();
  } catch (e) {
    alert('Could not reach the server.');
  }
}

function showError(msg) {
  const el = document.getElementById('errorBanner');
  el.textContent = msg;
  el.style.display = 'block';
}
function hideError() {
  document.getElementById('errorBanner').style.display = 'none';
}
function esc(s) {
  return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}

// Submit on Enter in any input
document.querySelectorAll('input').forEach(inp => {
  inp.addEventListener('keydown', e => { if (e.key === 'Enter') addCard(); });
});

// Delegated remove handler — survives innerHTML replacement
document.getElementById('listContainer').addEventListener('click', e => {
  const btn = e.target.closest('button[data-card]');
  if (btn) removeCard(btn.dataset.card);
});

loadCards();
</script>
</body>
</html>`
