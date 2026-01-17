package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/SteliosSpanos/mini-CAS/pkg/catalog"
	"github.com/SteliosSpanos/mini-CAS/pkg/client"
)

func Status() {
	c, err := client.NewClientFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create client: %v\n", err)
		os.Exit(1)
	}
	defer c.Close()

	entries, err := c.GetCatalog(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "No files tracked in catalog\n")
		os.Exit(1)
	}

	totalEntries := len(entries)

	uniqueHashes := make(map[string]bool)
	totalSize := uint64(0)

	for _, entry := range entries {
		uniqueHashes[entry.Hash] = true
		totalSize += entry.Filesize
	}

	uniqueBlobs := len(uniqueHashes)

	actualStorage := uint64(0)
	seenHashes := make(map[string]bool)

	for _, entry := range entries {
		if !seenHashes[entry.Hash] {
			actualStorage += entry.Filesize
			seenHashes[entry.Hash] = true
		}
	}

	spaceSaved := totalSize - actualStorage
	percentageSaved := 0.0
	if totalSize > 0 {
		percentageSaved = float64(spaceSaved) / float64(totalSize) * 100
	}

	fmt.Println("Repository Statistics:")
	fmt.Println("======================")
	fmt.Printf("Files Tracked: %d\n", totalEntries)
	fmt.Printf("Unique Blobs: %d\n", uniqueBlobs)
	fmt.Printf("Total File Size: %s\n", catalog.FormatSize(totalSize))
	fmt.Printf("Actual Storage: %s\n", catalog.FormatSize(actualStorage))
	fmt.Printf("Space Saved: %s (%.1f%%)\n", catalog.FormatSize(spaceSaved), percentageSaved)

}
