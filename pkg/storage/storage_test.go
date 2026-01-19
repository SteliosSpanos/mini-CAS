package storage

import (
	"os"
	"path/filepath"
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
