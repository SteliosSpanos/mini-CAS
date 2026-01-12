package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

type BlobResponse struct {
	Hash   string `json:"hash"`
	Size   int64  `json:"size"`
	Exists bool   `json:"exists,omitempty"`
}

type HealthResponse struct {
	Status      string `json:"status"`
	TotalFiles  int    `json:"total_files"`
	UniqueBlobs int    `json:"unique_blobs"`
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return fmt.Errorf("failed to encode json: %w", err)
	}

	return nil
}

func WriteError(w http.ResponseWriter, status int, message string) {
	response := ErrorResponse{
		Error:   http.StatusText(status),
		Code:    status,
		Message: message,
	}

	WriteJSON(w, status, response)
}

func WriteBlob(w http.ResponseWriter, hash string, size int64, reader io.ReadCloser) error {
	defer reader.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, hash))
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

	if _, err := io.Copy(w, reader); err != nil {
		return fmt.Errorf("failed to stream blob: %w", err)
	}

	return nil
}
