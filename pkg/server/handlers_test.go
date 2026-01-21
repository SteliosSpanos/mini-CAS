package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SteliosSpanos/mini-CAS/pkg/catalog"
	"github.com/SteliosSpanos/mini-CAS/pkg/storage"
)

func setupTestServer(t *testing.T) *Server {
	t.Helper()

	casDir := t.TempDir()
	storageDir := filepath.Join(casDir, "storage")

	if err := os.MkdirAll(storageDir, 0755); err != nil {
		t.Fatalf("failed to create storage dir: %v", err)
	}

	cat := catalog.NewCatalog(casDir)
	cat.Load()

	t.Cleanup(func() {
		cat.Close()
	})

	return &Server{
		config: Config{
			AuthToken:   "test-token",
			CORSOrigins: []string{"*"},
		},
		catalog: cat,
		casDir:  casDir,
		logger:  log.New(io.Discard, "", 0),
	}
}

func validTestHash() string {
	return "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
}

func TestHandleHealth(t *testing.T) {
	server := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	server.handleHealth(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var response HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("status = %q, want %q", response.Status, "ok")
	}
}

func TestHandleGetBlob_InvalidHash(t *testing.T) {
	server := setupTestServer(t)

	testCases := []struct {
		name string
		hash string
	}{
		{"too short", "abc123"},
		{"too long", "abcdef1234567890abcdef1234567890abcdef1234567890abcdef12345678901"},
		{"invalid chars", "ghijkl1234567890abcdef1234567890abcdef1234567890abcdef1234567"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/blob/"+tc.hash, nil)
			req.SetPathValue("hash", tc.hash)
			rec := httptest.NewRecorder()

			server.handleGetBlob(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestHandleGetBlob_NotFound(t *testing.T) {
	server := setupTestServer(t)

	hash := validTestHash()
	req := httptest.NewRequest(http.MethodGet, "/blob/"+hash, nil)
	req.SetPathValue("hash", hash)
	rec := httptest.NewRecorder()

	server.handleGetBlob(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestHandlePostBlob(t *testing.T) {
	server := setupTestServer(t)

	content := "hello server test"
	body := strings.NewReader(content)

	req := httptest.NewRequest(http.MethodPost, "/blobs", body)
	rec := httptest.NewRecorder()

	server.handlePostBlob(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var response BlobResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Hash) != 64 {
		t.Errorf("hash length = %d, want 64", len(response.Hash))
	}

	reader, err := storage.OpenBlob(server.casDir, response.Hash)
	if err != nil {
		t.Fatalf("blob was not stored: %v", err)
	}
	defer reader.Close()
}

func TestHandleGetBlob_Success(t *testing.T) {
	server := setupTestServer(t)

	content := "content for download test"
	hash, err := storage.WriteBlobStream(server.casDir, strings.NewReader(content))
	if err != nil {
		t.Fatalf("failed to write blob: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/blob/"+hash, nil)
	req.SetPathValue("hash", hash)
	rec := httptest.NewRecorder()

	server.handleGetBlob(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if rec.Body.String() != content {
		t.Errorf("body = %q, want %q", rec.Body.String(), content)
	}
}
