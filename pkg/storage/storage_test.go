package storage

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SteliosSpanos/mini-CAS/pkg/objects"
)

func TestWriteBlob(t *testing.T) {
	casDir := t.TempDir()

	content := []byte("hello, this is the mini-CAS")
	blob := objects.NewBlob(content)

	hash, err := WriteBlob(casDir, *blob)

	if err != nil {
		t.Fatalf("WriteBlob() error: %v", err)
	}

	if len(hash) != 64 {
		t.Errorf("hash length = %d, want 64", len(hash))
	}

	expectedPath := filepath.Join(casDir, "storage", hash[:2], hash[2:4], hash)
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("file not created at %s", expectedPath)
	}

	storedData, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(storedData) != string(content) {
		t.Errorf("stored content = %q, want %q", storedData, content)
	}
}

func TestWriteBlob_Deduplication(t *testing.T) {
	casDir := t.TempDir()

	content := []byte("duplicate")
	blob := objects.NewBlob(content)

	hash1, err := WriteBlob(casDir, *blob)
	if err != nil {
		t.Fatalf("first WriteBlob() error: %v", err)
	}

	blobPath := filepath.Join(casDir, "storage", hash1[:2], hash1[2:4], hash1)
	stat1, err := os.Stat(blobPath)
	if err != nil {
		t.Fatalf("os.Stat() error: %v", err)
	}
	modTime1 := stat1.ModTime()

	hash2, err := WriteBlob(casDir, *blob)
	if err != nil {
		t.Fatalf("second WriteBlob() error: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("hash mismatch: first=%s, second=%s", hash1, hash2)
	}

	stat2, err := os.Stat(blobPath)
	if err != nil {
		t.Fatalf("os.Stat() error: %v", err)
	}
	modTime2 := stat2.ModTime()

	if !modTime1.Equal(modTime2) {
		t.Errorf("file was rewritten: modTime1=%v, modTime2=%v", modTime1, modTime2)
	}
}

func TestWriteBlobStream(t *testing.T) {
	casDir := t.TempDir()

	os.MkdirAll(filepath.Join(casDir, "storage"), 0755)

	content := "streaming content test"
	reader := strings.NewReader(content)

	hash, err := WriteBlobStream(casDir, reader)
	if err != nil {
		t.Fatalf("WriteBlobStream() error: %v", err)
	}

	if len(hash) != 64 {
		t.Errorf("hash length = %d, want 64", len(hash))
	}

	blobPath := filepath.Join(casDir, "storage", hash[:2], hash[2:4], hash)
	data, err := os.ReadFile(blobPath)
	if err != nil {
		t.Fatalf("failed to read blob: %v", err)
	}

	if string(data) != content {
		t.Errorf("content = %q, want %q", data, content)
	}

	info, _ := os.Stat(blobPath)
	if info.Mode().Perm() != 0444 {
		t.Errorf("permissions = %o, want 0444", info.Mode().Perm())
	}
}

func TestReadBlob(t *testing.T) {
	casDir := t.TempDir()
	content := []byte("data to be read back")
	blob := objects.NewBlob(content)

	hash, err := WriteBlob(casDir, *blob)
	if err != nil {
		t.Fatalf("setup: WriteBlob() error: %v", err)
	}

	data, err := ReadBlob(casDir, hash)
	if err != nil {
		t.Fatalf("ReadBlob() error: %v", err)
	}

	if string(data) != string(content) {
		t.Errorf("ReadBlob() = %q, want %q", data, content)
	}
}

func TestReadBlob_NotFound(t *testing.T) {
	casDir := t.TempDir()

	os.MkdirAll(filepath.Join(casDir, "storage"), 0755)

	fakeHash := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	_, err := ReadBlob(casDir, fakeHash)

	if err == nil {
		t.Error("ReadBlob() expected error for non-existent blob, got nil")
	}
}

func TestLoadBlob(t *testing.T) {
	casDir := t.TempDir()

	content := []byte("content for LoadBlob test")
	blob := objects.NewBlob(content)

	hash, err := WriteBlob(casDir, *blob)
	if err != nil {
		t.Fatalf("setup: WriteBlob() error: %v", err)
	}

	loadedBlob, err := LoadBlob(casDir, hash)
	if err != nil {
		t.Fatalf("LoadBlob() error: %v", err)
	}

	if loadedBlob == nil {
		t.Fatalf("LoadBlob() returned nil blob")
	}

	if string(loadedBlob.Data) != string(content) {
		t.Errorf("LoadBlob().Data = %q, want %q", loadedBlob.Data, content)
	}
}

func TestOpenBlob(t *testing.T) {
	casDir := t.TempDir()

	content := []byte("content for OpenBlob test")
	blob := objects.NewBlob(content)

	hash, err := WriteBlob(casDir, *blob)
	if err != nil {
		t.Fatalf("setup: WriteBlob() error: %v", err)
	}

	reader, err := OpenBlob(casDir, hash)
	if err != nil {
		t.Fatalf("OpenBlob() error: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("io.ReadAll() error: %v", err)
	}

	if string(data) != string(content) {
		t.Errorf("OpenBlob() content = %q, want %q", data, content)
	}
}
