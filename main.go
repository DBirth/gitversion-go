package main

import (
	"fmt"
	"log"

	"github.com/go-git/go-git/v5"
)

func main() {
	// Open the repository in the current directory.
	r, err := git.PlainOpen(".")
	if err != nil {
		log.Fatalf("failed to open repository: %v", err)
	}

	// Get the reference to HEAD.
	head, err := r.Head()
	if err != nil {
		log.Fatalf("failed to get HEAD: %v", err)
	}

	// Print the hash of the HEAD commit.
	fmt.Println(head.Hash())
}
