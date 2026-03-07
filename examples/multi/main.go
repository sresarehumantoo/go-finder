// Multi example shows how to use go-finder for multi-file selection.
// Selections persist across directory navigation.
//
// Usage:
//
//	go run ./examples/multi
package main

import (
	"fmt"
	"os"

	finder "github.com/SREsAreHumanToo/go-finder"
)

func main() {
	paths, err := finder.PickMultiple(
		finder.WithTitle("Select files to process"),
		finder.WithFilter("*.go", "*.md", "*.txt"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Selected %d file(s):\n", len(paths))
	for _, p := range paths {
		fmt.Println("  ", p)
	}
}
