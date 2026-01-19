package catalog

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Entry struct {
	Filepath string    `json:"filepath"`
	Hash     string    `json:"hash"`
	Filesize uint64    `json:"file_size"`
	ModTime  time.Time `json:"modification_time"`
}

type Catalog struct {
	db     *sql.DB
	casDir string
}

func NewCatalog(casDir string) *Catalog {
	return &Catalog{
		casDir: casDir,
	}
}

func (c *Catalog) init() error {
	if c.db != nil {
		return nil
	}

	dbPath := filepath.Join(c.casDir, "catalog.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA busy_timeout=5000",
	}

	for _, p := range pragmas {
		db.Exec(p)
	}

	schema := `
			CREATE TABLE IF NOT EXISTS entries (
					filepath TEXT PRIMARY KEY NOT NULL,
					hash TEXT NOT NULL,
					filesize INTEGER NOT NULL,
					modtime INTEGER NOT NULL
			);
			CREATE INDEX IF NOT EXISTS idx_hash ON entries(hash);
	`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return fmt.Errorf("failed to create schema: %w", err)
	}

	c.db = db
	return nil
}

func (c *Catalog) AddEntry(entry Entry) error {
	if err := c.init(); err != nil {
		return err
	}

	query := `
			INSERT INTO entries (filepath, hash, filesize, modTime)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(filepath) DO UPDATE SET
					hash = excluded.hash,
					filesize = excluded.filesize,
					modTime = excluded.modTime
	`

	_, err := c.db.Exec(query, entry.Filepath, entry.Hash, entry.Filesize, entry.ModTime.UnixNano())
	return err
}

func (c *Catalog) GetEntry(path string) (Entry, error) {
	if err := c.init(); err != nil {
		return Entry{}, err
	}

	var entry Entry
	var modtime int64

	err := c.db.QueryRow(
		"SELECT filepath, hash, filesize, modtime FROM entries WHERE filepath = ?",
		path,
	).Scan(&entry.Filepath, &entry.Hash, &entry.Filesize, &modtime)

	if err == sql.ErrNoRows {
		return Entry{}, fmt.Errorf("path not found in catalog: %s", path)
	}

	if err != nil {
		return Entry{}, err
	}

	entry.ModTime = time.Unix(0, modtime)
	return entry, nil
}

func (c *Catalog) ListEntries() ([]Entry, error) {
	if err := c.init(); err != nil {
		return nil, err
	}

	rows, err := c.db.Query("SELECT filepath, hash, filesize, modtime FROM entries ORDER BY filepath")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var entry Entry
		var modtime int64

		if err := rows.Scan(&entry.Filepath, &entry.Hash, &entry.Filesize, &modtime); err != nil {
			return nil, err
		}

		entry.ModTime = time.Unix(0, modtime)
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

func (c *Catalog) Save() error {
	return c.init()
}

func (c *Catalog) Load() error {
	return c.init()
}

func (c *Catalog) WriteTo(w io.Writer) (int64, error) {
	entries, err := c.ListEntries()
	if err != nil {
		return 0, fmt.Errorf("failed to list entries: %w", err)
	}

	data, err := json.MarshalIndent(entries, "", " ")
	if err != nil {
		return 0, fmt.Errorf("failed to marshal entries: %w", err)
	}

	n, err := w.Write(data)
	if err != nil {
		return 0, fmt.Errorf("failed to write into Writer: %w", err)
	}

	return int64(n), nil
}

func (c *Catalog) ReadFrom(r io.Reader) (int64, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return 0, fmt.Errorf("failed to read from Reader: %w", err)
	}

	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		var entryMap map[string]Entry

		if err := json.Unmarshal(data, &entryMap); err != nil {
			return 0, fmt.Errorf("failed to unmarshal data: %w", err)
		}

		for _, entry := range entries {
			entries = append(entries, entry)
		}
	}

	if err := c.init(); err != nil {
		return 0, err
	}

	for _, entry := range entries {
		if err := c.AddEntry(entry); err != nil {
			return 0, fmt.Errorf("failed to add entry: %w", err)
		}
	}

	return int64(len(data)), nil
}

func (c *Catalog) Close() error {
	if c.db != nil {
		return c.db.Close()
	}

	return nil
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
