package commands

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/SteliosSpanos/mini-CAS/pkg/catalog"
	"github.com/SteliosSpanos/mini-CAS/pkg/path"
	"github.com/SteliosSpanos/mini-CAS/pkg/storage"
)

func Verify() {
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

	totalFiles := len(entries)
	verified := 0
	corrupted := 0
	missing := 0

	for _, entry := range entries {
		file, err := storage.OpenBlob(repo.RootDir, entry.Hash)
		if err != nil {
			fmt.Printf(" MISSING: %s (hash %s)\n", entry.Filepath, entry.Hash[:8])
			missing++
			continue
		}

		hasher := sha256.New()

		if _, err := io.Copy(hasher, file); err != nil {
			fmt.Printf(" ERROR: %s - failed to read %v\n", entry.Filepath, err)
			file.Close()
			corrupted++
			continue
		}
		file.Close()

		computedHash := fmt.Sprintf("%x", hasher.Sum(nil))

		if computedHash != entry.Hash {
			fmt.Printf(" CORRUPT: %s\n", entry.Filepath)
			fmt.Printf("  Expected: %s\n", entry.Hash[:8])
			fmt.Printf("  Got:      %s\n", entry.Hash[:8])
			corrupted++
		} else {
			fmt.Printf("  OK: %s\n", entry.Filepath)
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
