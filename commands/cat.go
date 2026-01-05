package commands

import (
	"fmt"
	"os"

	"github.com/SteliosSpanos/mini-CAS/pkg/catalog"
	"github.com/SteliosSpanos/mini-CAS/pkg/path"
)

func Cat(args []string) {
	repo, err := path.Open(".")
	if err != nil {
		fmt.Println("Not a CAS repository. Run './cas init' first")
		os.Exit(1)
	}

	if len(args) != 1 {
		fmt.Println("Usage: ./cas cat <filepath>")
		os.Exit(1)
	}

	filePath := args[0]

	cat := catalog.NewCatalog(repo.RootDir)
	if err := cat.Load(); err != nil {
		fmt.Printf("Failed to load catalog: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		fmt.Printf("Failed to access: %v\n", err)
		os.Exit(1)
	}

	if info.IsDir() {
		fmt.Println("Path should be a file")
		os.Exit(1)
	}

	entry, err := cat.GetEntry(filePath)
	if err != nil {
		fmt.Printf("This file doesn't exist in the catalog: %v\n", err)
		os.Exit(1)
	}

	hashShort := entry.Hash[:8]
	sizeStr := formatSize(entry.Filesize)
	modTime := entry.ModTime.Format("2006-01-02 15:04")

	fmt.Printf("%-50s %-10s %-12s %s\n", entry.Filepath, hashShort, sizeStr, modTime)
}
