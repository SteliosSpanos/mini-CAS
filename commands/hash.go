package commands

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

func HashFile(args []string) {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: ./cas <filename>\n")
		os.Exit(1)
	}

	fileName := args[0]

	file, err := os.Open(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open %s: %v\n", fileName, err)
		os.Exit(1)
	}
	defer file.Close()

	hasher := sha256.New()
	io.Copy(hasher, file)

	fmt.Printf("%x\n", hasher.Sum(nil))
}
