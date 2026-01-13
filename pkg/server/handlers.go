package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/SteliosSpanos/mini-CAS/pkg/storage"
)

var hashRegex = regexp.MustCompile("^[a-f0-9]{64}$")

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	entries := s.catalog.ListEntries()

	uniqueHashes := make(map[string]bool)
	for _, entry := range entries {
		uniqueHashes[entry.Hash] = true
	}

	response := HealthResponse{
		Status:      "ok",
		TotalFiles:  len(entries),
		UniqueBlobs: len(uniqueHashes),
	}

	WriteJSON(w, http.StatusOK, response)
}

func (s *Server) handleGetBlob(w http.ResponseWriter, r *http.Request) {
	hash := r.PathValue("hash")

	if !isValidHash(hash) {
		WriteError(w, http.StatusBadRequest, "Invalid hash format: must be 64 hex characters")
		return
	}

	reader, err := storage.OpenBlob(s.casDir, hash)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteError(w, http.StatusNotFound, "Blob not found")
		} else {
			s.logger.Printf("Error opening blob %s: %v", hash, err)
			WriteError(w, http.StatusInternalServerError, "Failed to read blob")
		}

		return
	}
	defer reader.Close()

	size := getSizeFromFile(reader)

	if r.Method == http.MethodHead {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
		w.Header().Set("ETag", fmt.Sprintf(`"%s"`, hash))
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		reader.Close()
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := WriteBlob(w, hash, size, reader); err != nil {
		s.logger.Printf("Error streaming blob %s: %v", hash, err)
	}
}

func (s *Server) handleStatBlob(w http.ResponseWriter, r *http.Request) {
	hash := r.PathValue("hash")

	if !isValidHash(hash) {
		WriteError(w, http.StatusBadRequest, "Invalid hash format")
		return
	}

	reader, err := storage.OpenBlob(s.casDir, hash)
	if err != nil {
		WriteJSON(w, http.StatusOK, BlobResponse{
			Hash:   hash,
			Size:   0,
			Exists: false,
		})
		return
	}
	defer reader.Close()

	size := getSizeFromFile(reader)

	WriteJSON(w, http.StatusOK, BlobResponse{
		Hash:   hash,
		Size:   size,
		Exists: true,
	})
}

func (s *Server) handleGetCatalog(w http.ResponseWriter, r *http.Request) {
	if err := s.catalog.Load(); err != nil {
		s.logger.Printf("Error reloading catalog: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to load catalog")
		return
	}

	if filepath := r.URL.Query().Get("filepath"); filepath != "" {
		entry, err := s.catalog.GetEntry(filepath)
		if err != nil {
			WriteError(w, http.StatusNotFound, "Entry not found")
			return
		}
		WriteJSON(w, http.StatusOK, entry)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := s.catalog.WriteTo(w); err != nil {
		s.logger.Printf("Error writing catalog: %v", err)
	}
}

func (s *Server) handlePostBlob(w http.ResponseWriter, r *http.Request) {
	hash, err := storage.WriteBlobStream(s.casDir, r.Body)
	if err != nil {
		s.logger.Printf("Error writing blob: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to write blob")
		return
	}

	reader, err := storage.OpenBlob(s.casDir, hash)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to read blob after write")
		return
	}
	defer reader.Close()

	size := getSizeFromFile(reader)

	response := BlobResponse{
		Hash: hash,
		Size: size,
	}

	WriteJSON(w, http.StatusCreated, response)
}

func getSizeFromFile(reader io.ReadCloser) int64 {
	var size int64
	if file, ok := reader.(*os.File); ok {
		if stat, err := file.Stat(); err == nil {
			size = stat.Size()
		}
	}
	return size
}

func isValidHash(hash string) bool {
	return hashRegex.MatchString(hash)
}
