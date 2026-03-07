// Folder example shows how to use go-finder as a directory picker.
//
// Usage:
//
//	go run ./examples/folder
package main

import (
	"fmt"
	"os"

	finder "github.com/SREsAreHumanToo/go-finder"
)

func main() {
	dir, err := finder.PickFolder(
		finder.WithTitle("Choose a project directory"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Selected directory:", dir)
}
