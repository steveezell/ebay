package wantlist

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Card struct {
	Name      string    `yaml:"name"      json:"name"`
	Query     string    `yaml:"query"     json:"query"`
	MaxPrice  float64   `yaml:"max_price" json:"max_price"`
	Condition string    `yaml:"condition,omitempty" json:"condition"`
	Notes     string    `yaml:"notes,omitempty"     json:"notes"`
	Added     time.Time `yaml:"added"     json:"added"`
}

type WantList struct {
	Cards []Card `yaml:"cards"`
	path  string
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".card-hunt", "wantlist.yaml")
}

func Load() (*WantList, error) {
	path := DefaultPath()
	wl := &WantList{path: path}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return wl, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read wantlist: %w", err)
	}
	if err := yaml.Unmarshal(data, wl); err != nil {
		return nil, fmt.Errorf("parse wantlist: %w", err)
	}
	return wl, nil
}

func (w *WantList) Save() error {
	if err := os.MkdirAll(filepath.Dir(w.path), 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(w)
	if err != nil {
		return err
	}
	return os.WriteFile(w.path, data, 0644)
}

func (w *WantList) Add(card Card) error {
	for _, c := range w.Cards {
		if strings.EqualFold(c.Name, card.Name) {
			return fmt.Errorf("card already on want list: %s", card.Name)
		}
	}
	w.Cards = append(w.Cards, card)
	return w.Save()
}

// Remove removes the first card whose name contains the given substring (case-insensitive).
// Returns the removed card name and true on success.
func (w *WantList) Remove(nameSubstr string) (string, bool) {
	lower := strings.ToLower(nameSubstr)
	for i, c := range w.Cards {
		if strings.Contains(strings.ToLower(c.Name), lower) {
			name := c.Name
			w.Cards = append(w.Cards[:i], w.Cards[i+1:]...)
			_ = w.Save()
			return name, true
		}
	}
	return "", false
}
