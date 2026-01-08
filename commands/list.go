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
		fmt.Fprintf(os.Stderr, "Not a CAS repository. Run './cas init' first: %v\n", err)
		os.Exit(1)
	}

	cat := catalog.NewCatalog(repo.RootDir)
	if err := cat.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load catalog: %v\n", err)
		os.Exit(1)
	}

	entries := cat.ListEntries()

	if len(entries) == 0 {
		fmt.Fprintf(os.Stderr, "No files tracked in catalog\n")
		os.Exit(1)
	}

	fmt.Printf("%-50s %-10s %-12s %s\n", "FILEPATH", "HASH", "SIZE", "MODIFIED")
	fmt.Println("====================================================================================================")

	for _, entry := range entries {
		hashShort := entry.Hash[:8]
		sizeStr := catalog.FormatSize(entry.Filesize)
		modTime := entry.ModTime.Format("2006-01-02 15:04")

		fmt.Printf("%-50s %-10s %-12s  %s\n", entry.Filepath, hashShort, sizeStr, modTime)
	}

	fmt.Printf("\nTotal files: %d\n", len(entries))
}
