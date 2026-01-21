package client

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestClient(t *testing.T) (*LocalClient, string) {
	t.Helper()

	casDir := t.TempDir()

	storageDir := filepath.Join(casDir, "storage")
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		t.Fatalf("failed to create storage fir: %v", err)
	}

	client, err := NewLocalClient(casDir)
	if err != nil {
		t.Fatalf("NewLocalClient() error: %v", err)
	}

	return client, casDir
}

func validHash() string {
	return "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
}

func TestLocalClientImplementsClient(t *testing.T) {
	var _ Client = (*LocalClient)(nil)
}

func TestUpload_Download_RoundTrip(t *testing.T) {
	client, _ := setupTestClient(t)
	defer client.Close()

	content := "hello mini-CAS"
	ctx := context.Background()

	hash, err := client.Upload(ctx, strings.NewReader(content))
	if err != nil {
		t.Fatalf("Upload() error: %v", err)
	}

	if len(hash) != 64 {
		t.Errorf("hash length = %d, want 64", len(hash))
	}

	reader, err := client.Download(ctx, hash)
	if err != nil {
		t.Fatalf("Download() error: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("io.ReadAll() error: %v", err)
	}

	if string(data) != content {
		t.Errorf("downloaded = %q, want %q", data, content)
	}
}

func TestDownload_InvalidHash(t *testing.T) {
	client, _ := setupTestClient(t)
	defer client.Close()

	ctx := context.Background()

	invalidHashes := []struct {
		name string
		hash string
	}{
		{"empty", ""},
		{"too short", "abc123"},
		{"63 chars", "abcdef1234567890abcdef1234567890abcdef1234567890abcdef123456789"},
	}

	for _, tc := range invalidHashes {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.Download(ctx, tc.hash)

			if !errors.Is(err, ErrInvalidHash) {
				t.Errorf("Download(%q) error = %v, want ErrInvalidHash", tc.hash, err)
			}
		})
	}
}

func TestDownload_BlobNotFound(t *testing.T) {
	client, _ := setupTestClient(t)
	defer client.Close()

	ctx := context.Background()

	_, err := client.Download(ctx, validHash())

	if !errors.Is(err, ErrBlobNotFound) {
		t.Errorf("Download() error = %v, want ErrBlobNotFound", err)
	}
}
