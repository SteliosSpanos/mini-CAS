package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/SteliosSpanos/mini-CAS/pkg/client"
)

func Cat(args []string) {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: ./cas cat <filepath>\n")
		os.Exit(1)
	}

	filePath := args[0]

	c, err := client.NewClientFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create client: %v\n", err)
		os.Exit(1)
	}
	defer c.Close()

	ctx := context.Background()

	entry, err := c.GetEntry(ctx, filePath)
	if err != nil {
		if errors.Is(err, client.ErrEntryNotFound) {
			fmt.Fprintf(os.Stderr, "This file doesn't exist in the catalog: %s\n", filePath)
		} else {
			fmt.Fprintf(os.Stderr, "Failed to get entry: %v\n", err)
		}
		os.Exit(1)
	}

	reader, err := c.Download(ctx, entry.Hash)
	if err != nil {
		if errors.Is(err, client.ErrBlobNotFound) {
			fmt.Fprintf(os.Stderr, "Blob not found in storage: %s\n", entry.Hash)
		} else {
			fmt.Fprintf(os.Stderr, "Failed to download blob: %v\n", err)
		}
		os.Exit(1)
	}
	defer reader.Close()

	io.Copy(os.Stdout, reader)
}
