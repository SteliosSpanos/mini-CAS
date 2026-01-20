package catalog

import (
	"testing"
	"time"
)

func TestAddEntry_GetEntry(t *testing.T) {
	casDir := t.TempDir()
	cat := NewCatalog(casDir)
	defer cat.Close()

	modTime := time.Now().Truncate(time.Microsecond)
	entry := Entry{
		Filepath: "docs/readme.txt",
		Hash:     "abc123def456abc123def456abc123def456abc123def456abc123def456abc1",
		Filesize: 1024,
		ModTime:  modTime,
	}

	err := cat.AddEntry(entry)
	if err != nil {
		t.Fatalf("AddEntry() error: %v", err)
	}

	got, err := cat.GetEntry("docs/readme.txt")
	if err != nil {
		t.Fatalf("GetEntry() error: %v", err)
	}

	if got.Filepath != entry.Filepath {
		t.Errorf("Filepath = %q, want %q", got.Filepath, entry.Filepath)
	}

	if got.Hash != entry.Hash {
		t.Errorf("Hash  = %q, want %q", got.Hash, entry.Hash)
	}

	if got.Filesize != entry.Filesize {
		t.Errorf("Filesize = %d, want %d", got.Filesize, entry.Filesize)
	}
}

func TestGetEntry_NotFound(t *testing.T) {
	casDir := t.TempDir()
	cat := NewCatalog(casDir)
	defer cat.Close()

	_, err := cat.GetEntry("nonexistent/file.txt")

	if err == nil {
		t.Error("GetEntry() expected error for non-existent path, got nil")
	}
}

func TestAddEntry_Upset(t *testing.T) {
	casDir := t.TempDir()
	cat := NewCatalog(casDir)
	defer cat.Close()

	path := "myfile.txt"

	entry1 := Entry{
		Filepath: path,
		Hash:     "aaaa1111aaaa1111aaaa1111aaaa1111aaaa1111aaaa1111aaaa1111aaaa1111",
		Filesize: 100,
		ModTime:  time.Now().Truncate(time.Microsecond),
	}
	cat.AddEntry(entry1)

	entry2 := Entry{
		Filepath: path,
		Hash:     "bbbb2222bbbb2222bbbb2222bbbb2222bbbb2222bbbb2222bbbb2222bbbb2222",
		Filesize: 200,
		ModTime:  time.Now().Truncate(time.Microsecond),
	}
	cat.AddEntry(entry2)

	got, err := cat.GetEntry(path)
	if err != nil {
		t.Fatalf("GetEntry() error: %v", err)
	}

	if got.Hash != entry2.Hash {
		t.Errorf("Hash = %q, want %q (should be updated)", got.Hash, entry2.Hash)
	}

	if got.Filesize != entry2.Filesize {
		t.Errorf("Filesize = %d, want %d", got.Filesize, entry2.Filesize)
	}

	entries, _ := cat.ListEntries()
	if len(entries) != 1 {
		t.Errorf("ListEntries() = %d entries, want 1", len(entries))
	}
}
