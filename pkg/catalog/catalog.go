package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Entry struct {
	Filepath string    `json:"filepath"`
	Hash     string    `json:"hash"`
	Filesize uint64    `json:"file_size"`
	ModTime  time.Time `json:"modification_time"`
}

type Catalog struct {
	entries map[string]Entry
	casDir  string
}

func NewCatalog(casDir string) *Catalog {
	return &Catalog{
		casDir:  casDir,
		entries: make(map[string]Entry),
	}
}

func (c *Catalog) AddEntry(entry Entry) {
	c.entries[entry.Filepath] = entry
}

func (c *Catalog) GetEntry(path string) (Entry, error) {
	entry, exists := c.entries[path]
	if !exists {
		return Entry{}, fmt.Errorf("path not found in catalog: %s", path)
	}

	return entry, nil
}

func (c *Catalog) ListEntries() []Entry {
	entries := make([]Entry, 0, len(c.entries))

	for _, entry := range c.entries {
		entries = append(entries, entry)
	}

	return entries
}

func (c *Catalog) Save() error {
	catalogPath := filepath.Join(c.casDir, "catalog.json")

	data, err := json.MarshalIndent(c.entries, "", " ")
	if err != nil {
		return fmt.Errorf("failed to turn to JSON: %w", err)
	}

	if err := os.WriteFile(catalogPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON to file: %w", err)
	}

	return nil
}

func (c *Catalog) Load() error {
	catalogPath := filepath.Join(c.casDir, "catalog.json")

	data, err := os.ReadFile(catalogPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to load JSON from file: %w", err)
	}

	if err := json.Unmarshal(data, &c.entries); err != nil {
		return fmt.Errorf("failed to turn JSON to catalog: %w", err)
	}

	return nil
}
