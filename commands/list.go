package commands

import (
	"fmt"
	"os"

	"github.com/SteliosSpanos/mini-CAS/pkg/catalog"
	"github.com/SteliosSpanos/mini-CAS/pkg/path"
)

func List() {
	repo, err := path.Open(".")
	if err != nil {
		fmt.Printf("Not a CAS repository. Run './cas init' first")
		os.Exit(1)
	}

	cat := catalog.NewCatalog(repo.RootDir)
	if err := cat.Load(); err != nil {
		fmt.Printf("Failed to load catalog: %v\n", err)
		os.Exit(1)
	}

	entries := cat.ListEntries()

	if len(entries) == 0 {
		fmt.Println("No files tracked in catalog")
		os.Exit(1)
	}

	fmt.Printf("%-50s %-10s %-12s %s\n", "FILEPATH", "HASH", "SIZE", "MODIFIED")
	fmt.Println("====================================================================================================")

	for _, entry := range entries {
		hashShort := entry.Hash[:8]
		sizeStr := formatSize(entry.Filesize)
		modTime := entry.ModTime.Format("2006-01-02 15:04")

		fmt.Printf("%-50s %-10s %-12s  %s\n", entry.Filepath, hashShort, sizeStr, modTime)
	}

	fmt.Printf("\nTotal files: %d\n", len(entries))
}

func formatSize(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
