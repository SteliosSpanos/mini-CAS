package commands

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/SteliosSpanos/mini-CAS/pkg/catalog"
	"github.com/SteliosSpanos/mini-CAS/pkg/path"
)

func Add(args []string) {
	repo, err := path.Open(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Not a CAS repository. Run './cas init' first: %v\n", err)
		os.Exit(1)
	}

	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: ./cas add <object>\n")
		os.Exit(1)
	}

	targetPath := args[0]

	cat := catalog.NewCatalog(repo.RootDir)
	if err := cat.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load catalog: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(targetPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to access %s: %v\n", targetPath, err)
		os.Exit(1)
	}

	if info.IsDir() {
		if err := addDirectory(repo.RootDir, targetPath, cat); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add directory: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := addFile(repo.RootDir, targetPath, cat); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add file: %v\n", err)
			os.Exit(1)
		}
	}

	if err := cat.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save catalog: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully added %s\n", targetPath)
}

func addDirectory(casDir, dirPath string, cat *catalog.Catalog) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk the new directory: %w", err)
		}

		if info.IsDir() {
			if filepath.Base(path) == ".cas" {
				return filepath.SkipDir
			}

			return nil
		}

		return addFile(casDir, path, cat)
	})
}

func addFile(casDir, filePath string, cat *catalog.Catalog) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %v\n", err)
	}

	tmpFile, err := os.CreateTemp(filepath.Join(casDir, "storage"), "tmp-")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)
	defer tmpFile.Close()

	hasher := sha256.New()
	multiWriter := io.MultiWriter(tmpFile, hasher)

	if _, err := io.Copy(multiWriter, file); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	tmpFile.Close()

	objectDir := filepath.Join(casDir, "storage", hash[:2], hash[2:4])
	objectPath := filepath.Join(objectDir, hash)

	if _, err := os.Stat(objectPath); err == nil {
		return nil
	}

	if err := os.MkdirAll(objectDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Rename is atomic so we only get the final file
	if err := os.Rename(tmpPath, objectPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	if err := os.Chmod(objectPath, 0444); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	entry := catalog.Entry{
		Filepath: filePath,
		Hash:     hash,
		Filesize: uint64(info.Size()),
		ModTime:  info.ModTime(),
	}

	cat.AddEntry(entry)

	fmt.Printf("     %s -> %s\n", filePath, hash[:8])
	return nil
}
