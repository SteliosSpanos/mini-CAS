package path

import (
	"fmt"
	"os"
	"path/filepath"
)

const CASDir = ".cas"

type Repository struct {
	RootDir string
}

func Init(path string) (*Repository, error) {
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current path: %w", err)
		}
	}

	casDir := filepath.Join(path, CASDir)

	if _, err := os.Stat(casDir); err == nil {
		return nil, fmt.Errorf("CAS repository already exists")
	}

	dirs := []string{
		casDir,
		filepath.Join(casDir, "storage"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return &Repository{RootDir: casDir}, nil
}

func Open(path string) (*Repository, error) {
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	casDir := filepath.Join(path, CASDir)

	if _, err := os.Stat(casDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("no CAS repository found: %w", err)
	}

	return &Repository{RootDir: casDir}, nil
}
