package catalog

import (
	"encoding/json"
	"fmt"
	"io"
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

	file, err := os.Create(catalogPath)
	if err != nil {
		return fmt.Errorf("failed to create catalog file: %w", err)
	}
	defer file.Close()

	_, err = c.WriteTo(file)
	return err
}

func (c *Catalog) Load() error {
	catalogPath := filepath.Join(c.casDir, "catalog.json")

	file, err := os.Open(catalogPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to open catalog: %w", err)
	}
	defer file.Close()

	_, err = c.ReadFrom(file)
	return err
}

func (c *Catalog) WriteTo(w io.Writer) (int64, error) {
	data, err := json.MarshalIndent(c.entries, "", " ")
	if err != nil {
		return 0, fmt.Errorf("failed to marshal catalog: %w", err)
	}

	n, err := w.Write(data)
	if err != nil {
		return 0, fmt.Errorf("failed to write catalog: %w", err)
	}

	return int64(n), nil
}

func (c *Catalog) ReadFrom(r io.Reader) (int64, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return 0, fmt.Errorf("failed to read catalog: %w", err)
	}

	if err := json.Unmarshal(data, &c.entries); err != nil {
		return 0, fmt.Errorf("failed to unmarshal catalog: %w", err)
	}

	return int64(len(data)), nil
}

func FormatSize(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
