package main

import (
	"fmt"
	"os"

	"github.com/SteliosSpanos/mini-CAS/commands"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./cas <command> [arguments]")
		fmt.Println("    init     Initialize CAS")
		fmt.Println("    hash     Displays the hash of a file for testing (CAS not needed)")
		fmt.Println("    add      Add file or directory in the storage")
		fmt.Println("    ls       List all the contents")
		fmt.Println("    cat      Show a specific file from the storage")
		fmt.Println("    status   Show CAS status and analytics")
		fmt.Println("    verify   Verify all the contents of the storage")
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
	case "status":
		commands.Status()
	case "hash":
		commands.HashFile(args)
	case "verify":
		commands.Verify()
	default:
		fmt.Println("Not a valid command")
		os.Exit(1)
	}
}
