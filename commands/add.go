package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SteliosSpanos/mini-CAS/pkg/catalog"
	"github.com/SteliosSpanos/mini-CAS/pkg/client"
)

func Add(args []string) {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: ./cas add <object>\n")
		os.Exit(1)
	}

	targetPath := args[0]

	c, err := client.NewClientFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create client: %v\n", err)
		os.Exit(1)
	}
	defer c.Close()

	info, err := os.Stat(targetPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to access %s: %v\n", targetPath, err)
		os.Exit(1)
	}

	ctx := context.Background()

	if info.IsDir() {
		if err := addDirectory(ctx, c, targetPath); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add directory: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := addFile(ctx, c, targetPath); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add file: %v\n", err)
			os.Exit(1)
		}
	}

	if err := c.SaveCatalog(ctx); err != nil {
		if errors.Is(err, client.ErrCatalogNotSupported) {
			fmt.Println("(Remote mode: server manages catalog)")
		} else {
			fmt.Fprintf(os.Stderr, "Failed to save catalog: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Successfully added %s\n", targetPath)
}

func addDirectory(ctx context.Context, c client.Client, dirPath string) error {
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

		return addFile(ctx, c, path)
	})
}

func addFile(ctx context.Context, c client.Client, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %v\n", err)
	}

	hash, err := c.Upload(ctx, file)
	if err != nil {
		return fmt.Errorf("failed to upload blob: %w", err)
	}

	entry := catalog.Entry{
		Filepath: filePath,
		Hash:     hash,
		Filesize: uint64(info.Size()),
		ModTime:  info.ModTime(),
	}

	if err := c.AddEntry(ctx, entry); err != nil {
		if errors.Is(err, client.ErrCatalogNotSupported) {
			fmt.Printf("     %s -> %s (uploaded)\n", filePath, hash[:8])
			return nil
		}
		return fmt.Errorf("failed to add catalog entry: %w", err)
	}

	fmt.Printf("     %s -> %s\n", filePath, hash[:8])
	return nil
}
