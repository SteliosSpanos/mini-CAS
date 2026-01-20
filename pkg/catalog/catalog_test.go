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

func TestListEntries(t *testing.T) {
	casDir := t.TempDir()
	cat := NewCatalog(casDir)
	defer cat.Close()

	entries := []Entry{
		{
			Filepath: "zebra.txt",
			Hash:     "aaaa1111aaaa1111aaaa1111aaaa1111aaaa1111aaaa1111aaaa1111aaaa1111",
			Filesize: 100,
			ModTime:  time.Now().Truncate(time.Microsecond),
		},
		{
			Filepath: "apple.txt",
			Hash:     "bbbb2222bbbb2222bbbb2222bbbb2222bbbb2222bbbb2222bbbb2222bbbb2222",
			Filesize: 200,
			ModTime:  time.Now().Truncate(time.Microsecond),
		},
		{
			Filepath: "mango.txt",
			Hash:     "cccc3333cccc3333cccc3333cccc3333cccc3333cccc3333cccc3333cccc3333",
			Filesize: 300,
			ModTime:  time.Now().Truncate(time.Microsecond),
		},
	}

	for _, entry := range entries {
		cat.AddEntry(entry)
	}

	got, err := cat.ListEntries()
	if err != nil {
		t.Fatalf("ListEntries() error: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("ListEntries() returned %d entries, want 3", len(got))
	}

	expectedOrder := []string{"apple.txt", "mango.txt", "zebra.txt"}
	for i, wantPath := range expectedOrder {
		if got[i].Filepath != wantPath {
			t.Errorf("ListEntries()[%d].Filepath = %q, want %q", i, got[i].Filepath, wantPath)
		}
	}
}

func TestListEntries_Empty(t *testing.T) {
	casDir := t.TempDir()
	cat := NewCatalog(casDir)
	defer cat.Close()

	got, err := cat.ListEntries()
	if err != nil {
		t.Fatalf("ListEntries() error: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("ListEntries() returned %d entries, want 0", len(got))
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name  string
		bytes uint64
		want  string
	}{
		{name: "zero bytes", bytes: 0, want: "0 B"},
		{name: "bytes", bytes: 512, want: "512B"},
		{name: "exactly 1 KB", bytes: 1024, want: "1.00 KB"},
		{name: "1.5 KB", bytes: 1536, want: "1.50 KB"},
		{name: "exactly 1 MB", bytes: 1024 * 1024, want: "1.00 MB"},
		{name: "exactly 1 GB", bytes: 1024 * 1024 * 1024, want: "1.00 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatSize(tt.bytes)
			if got != tt.want {
				t.Errorf("FormatSize(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}
