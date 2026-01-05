package commands

import (
	"fmt"
	"os"

	"github.com/SteliosSpanos/mini-CAS/pkg/path"
)

func Init() {
	repo, err := path.Init("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize CAS: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Initialized empty CAS in %s\n", repo.RootDir)
}
