package main

import (
	"fmt"
	"os"

	"github.com/SteliosSpanos/mini-CAS/commands"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./cas <command> [arguments]")
		fmt.Println("    init    Initialize CAS")
		fmt.Println("    add     Add file or directory in the storage")
		fmt.Println("    ls      List all the contents")
		fmt.Println("    cat     Show a specific file from the catalog")
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "init":
		commands.Init()
	case "add":
		commands.Add(args)
	case "ls":
		commands.List()
	case "cat":
		commands.Cat(args)
	default:
		fmt.Println("Not a valid command")
		os.Exit(1)
	}
}
