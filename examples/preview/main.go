// Preview example shows the optional preview pane, which previews the
// highlighted entry (a file's head, a directory's listing, or metadata)
// beside the file list.
//
// Usage:
//
//	go run ./examples/preview            # built-in preview
//	go run ./examples/preview -custom    # custom PreviewFunc
//
// The pane is hidden automatically on narrow terminals.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	finder "github.com/SREsAreHumanToo/go-finder"
)

func main() {
	custom := flag.Bool("custom", false, "use a custom PreviewFunc instead of the built-in preview")
	flag.Parse()

	opts := []finder.Option{
		finder.WithTitle("Pick a file (→/l to enter dirs, / to search)"),
		finder.WithPreview(true),
	}

	if *custom {
		// A PreviewFunc receives the highlighted entry and the pane size, and
		// returns the text to display. This one shows a small metadata card.
		opts = append(opts, finder.WithPreviewFunc(
			func(e finder.FileEntry, width, height int) string {
				kind := "file"
				if e.IsDir {
					kind = "directory"
				}
				return strings.Join([]string{
					"Name: " + e.Name,
					"Kind: " + kind,
					"Size: " + finder.FormatSize(e.Size),
					"Mode: " + e.Mode.String(),
				}, "\n")
			},
		))
	}

	path, err := finder.PickFile(opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Selected:", path)
}
