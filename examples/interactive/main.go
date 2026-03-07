// Interactive example shows how to use go-finder with file management
// actions enabled (create files/folders, delete entries).
//
// Usage:
//
//	go run ./examples/interactive
package main

import (
	"fmt"
	"os"

	finder "github.com/SREsAreHumanToo/go-finder"
)

func main() {
	path, err := finder.PickFile(
		finder.WithTitle("Manage files (n: new file, N: new folder, d: delete)"),
		finder.WithInteractive(true),
		finder.WithHidden(true),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Selected:", path)
}
