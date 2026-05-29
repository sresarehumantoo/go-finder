// Extensions example shows how to restrict the picker to a set of file
// extensions, the way many "attach a document" dialogs only allow .pdf,
// .docx, .doc, and so on.
//
// WithExtensions is opt-in: omit it and every file is shown (the default).
// Matching is case-insensitive, so "Report.PDF" is included by
// WithExtensions("pdf"). It also composes with WithFilter — a file is shown
// if it matches either the glob patterns or the extensions.
//
// Usage:
//
//	go run ./examples/extensions
//	go run ./examples/extensions pdf docx doc
package main

import (
	"fmt"
	"os"

	finder "github.com/rummage-dev/rummage"
)

func main() {
	// Allow the extensions to be overridden from the command line so you can
	// try different sets; defaults to common document types.
	exts := os.Args[1:]
	if len(exts) == 0 {
		exts = []string{"pdf", "docx", "doc"}
	}

	path, err := finder.PickFile(
		finder.WithTitle(fmt.Sprintf("Select a document (%v)", exts)),
		finder.WithExtensions(exts...),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Selected document:", path)
}
