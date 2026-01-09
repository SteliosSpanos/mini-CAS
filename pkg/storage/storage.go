package storage

import (
	"crypto/sha256"
	"fmt"
	"io"
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

func WriteBlobStream(casDir string, reader io.Reader) (string, error) {
	tmpFile, err := os.CreateTemp(filepath.Join(casDir, "storage"), "tmp-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)
	defer tmpFile.Close()

	hasher := sha256.New()
	multiWriter := io.MultiWriter(tmpFile, hasher)

	if _, err := io.Copy(multiWriter, reader); err != nil {
		return "", fmt.Errorf("failed to copy: %w", err)
	}

	hash := fmt.Sprintf("%x", hasher.Sum(nil))
	tmpFile.Close()

	objectDir := filepath.Join(casDir, "storage", hash[:2], hash[2:4])
	objectPath := filepath.Join(objectDir, hash)

	if _, err := os.Stat(objectPath); err == nil {
		return hash, nil
	}

	if err := os.MkdirAll(objectDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.Rename(tmpPath, objectPath); err != nil {
		return "", fmt.Errorf("failed to move file: %w", err)
	}

	if err := os.Chmod(objectPath, 0444); err != nil {
		return "", fmt.Errorf("failed to set permissions: %w", err)
	}

	return hash, nil
}
