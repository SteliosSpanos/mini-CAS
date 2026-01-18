package commands

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/SteliosSpanos/mini-CAS/pkg/client"
)

func Verify() {
	c, err := client.NewClientFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create client: %v\n", err)
		os.Exit(1)
	}
	defer c.Close()

	ctx := context.Background()

	entries, err := c.GetCatalog(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load catalog: %v\n", err)
		os.Exit(1)
	}

	totalFiles := len(entries)
	verified := 0
	corrupted := 0
	missing := 0

	for _, entry := range entries {
		reader, err := c.Download(ctx, entry.Hash)
		if err != nil {
			if errors.Is(err, client.ErrBlobNotFound) {
				fmt.Fprintf(os.Stderr, "MISSING: %s (hash %s)\n", entry.Filepath, entry.Hash)
				missing++
				continue
			}
			fmt.Printf("ERROR: %s - failed to download: %v\n", entry.Filepath, err)
			corrupted++
			continue
		}

		hasher := sha256.New()

		if _, err := io.Copy(hasher, reader); err != nil {
			fmt.Printf("ERROR: %s - failed to read %v\n", entry.Filepath, err)
			reader.Close()
			corrupted++
			continue
		}
		reader.Close()

		computedHash := fmt.Sprintf("%x", hasher.Sum(nil))

		if computedHash != entry.Hash {
			fmt.Printf("CORRUPT: %s\n", entry.Filepath)
			fmt.Printf(" Expected: %s\n", entry.Hash[:8])
			fmt.Printf(" Got:      %s\n", entry.Hash[:8])
			corrupted++
		} else {
			fmt.Printf("OK: %s\n", entry.Filepath)
			verified++
		}
	}

	fmt.Println("=====================")
	fmt.Println("Verification Results:")
	fmt.Printf("  Total Files: %d\n", totalFiles)
	fmt.Printf("  Verified Files: %d\n", verified)
	fmt.Printf("  Corrupted Files: %d\n", corrupted)

	if corrupted > 0 || missing > 0 {
		fmt.Printf("\nWARNING: Storage issues detected\n")
		os.Exit(1)
	} else {
		fmt.Printf("\nAll files verified successfully!\n")
	}
}
