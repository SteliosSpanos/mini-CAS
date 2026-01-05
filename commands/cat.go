package commands

import (
	"fmt"
	"os"

	"github.com/SteliosSpanos/mini-CAS/pkg/catalog"
	"github.com/SteliosSpanos/mini-CAS/pkg/path"
	"github.com/SteliosSpanos/mini-CAS/pkg/storage"
)

func Cat(args []string) {
	repo, err := path.Open(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Not a CAS repository. Run './cas init' first: %v\n", err)
		os.Exit(1)
	}

	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: ./cas cat <filepath>\n")
		os.Exit(1)
	}

	filePath := args[0]

	cat := catalog.NewCatalog(repo.RootDir)
	if err := cat.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load catalog: %v\n", err)
		os.Exit(1)
	}

	entry, err := cat.GetEntry(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "This file doesn't exist in the catalog: %v\n", err)
		os.Exit(1)
	}

	data, err := storage.ReadBlob(repo.RootDir, entry.Hash)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read blob from storage: %v\n", err)
		os.Exit(1)
	}

	os.Stdout.Write(data)
}
