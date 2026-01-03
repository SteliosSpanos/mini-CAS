package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SteliosSpanos/mini-CAS/pkg/objects"
)

func WriteBlob(casDir string, blob objects.Blob) (string, error) {
	hash := objects.Hash(blob)

	objectDir := filepath.Join(casDir, "storage", hash[:2], hash[2:4])
	objectPath := filepath.Join(objectDir, hash)

	if _, err := os.Stat(objectPath); err == nil { // Check if hash already exists to optimize I/O
		return hash, nil
	}

	if err := os.MkdirAll(objectDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create object directory: %w", err)
	}

	if err := os.WriteFile(objectPath, blob.Data, 0444); err != nil {
		return "", fmt.Errorf("failed to write object file: %w", err)
	}

	return hash, nil

}

func ReadBlob(casDir, hash string) ([]byte, error) {
	objectPath := filepath.Join(casDir, "storage", hash[:2], hash[2:4], hash)

	data, err := os.ReadFile(objectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read object: %w", err)
	}

	return data, nil
}

func LoadBlob(casDir, hash string) (*objects.Blob, error) {
	data, err := ReadBlob(casDir, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to read object: %w", err)
	}

	blob := objects.NewBlob(data)
	return blob, nil
}
