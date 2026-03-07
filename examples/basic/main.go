// Basic example demonstrates all go-finder picker modes with flag-based configuration.
//
// Usage:
//
//	go run ./examples/basic                          # Pick a single file
//	go run ./examples/basic -mode folder             # Pick a folder
//	go run ./examples/basic -mode any                # Pick a file or folder
//	go run ./examples/basic -mode multi              # Multi-select files
//	go run ./examples/basic -dir /tmp                # Start in /tmp
//	go run ./examples/basic -filter "*.go"           # Only show .go files
//	go run ./examples/basic -hidden                  # Show hidden files
//	go run ./examples/basic -interactive             # Enable create/delete actions
//	go run ./examples/basic -expand -dir ~/symlink   # Resolve symlinks to real paths
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	finder "github.com/SREsAreHumanToo/go-finder"
)

func main() {
	mode := flag.String("mode", "file", "Picker mode: file, folder, any, or multi")
	dir := flag.String("dir", "", "Starting directory (default: current directory)")
	filter := flag.String("filter", "", "Comma-separated glob filters (e.g. '*.go,*.txt')")
	hidden := flag.Bool("hidden", false, "Show hidden files")
	interactive := flag.Bool("interactive", false, "Enable create/delete actions (n: new file, N: new folder, d: delete)")
	expand := flag.Bool("expand", false, "Resolve symlinks to real paths")
	flag.Parse()

	var opts []finder.Option

	if *dir != "" {
		opts = append(opts, finder.WithStartDir(*dir))
	}

	if *filter != "" {
		patterns := strings.Split(*filter, ",")
		for i := range patterns {
			patterns[i] = strings.TrimSpace(patterns[i])
		}
		opts = append(opts, finder.WithFilter(patterns...))
	}

	if *hidden {
		opts = append(opts, finder.WithHidden(true))
	}

	if *interactive {
		opts = append(opts, finder.WithInteractive(true))
	}

	if *expand {
		opts = append(opts, finder.WithExpandSymlinks(true))
	}

	switch *mode {
	case "file":
		path, err := finder.PickFile(opts...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Selected:", path)

	case "folder":
		path, err := finder.PickFolder(opts...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Selected:", path)

	case "any":
		path, err := finder.PickAny(opts...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Selected:", path)

	case "multi":
		paths, err := finder.PickMultiple(opts...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Selected:")
		for _, p := range paths {
			fmt.Println("  ", p)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown mode: %s (use file, folder, any, or multi)\n", *mode)
		os.Exit(1)
	}
}
