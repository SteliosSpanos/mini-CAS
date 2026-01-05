package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SteliosSpanos/mini-CAS/pkg/catalog"
	"github.com/SteliosSpanos/mini-CAS/pkg/objects"
	"github.com/SteliosSpanos/mini-CAS/pkg/path"
	"github.com/SteliosSpanos/mini-CAS/pkg/storage"
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
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	blob := objects.NewBlob(data)

	hash, err := storage.WriteBlob(casDir, *blob)
	if err != nil {
		return fmt.Errorf("failed to write blob: %w", err)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %v\n", err)
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
