package finder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatchesAnyExtension(t *testing.T) {
	exts := []string{".pdf", ".docx", ".doc"}
	cases := []struct {
		name string
		want bool
	}{
		{"report.pdf", true},
		{"Report.PDF", true}, // case-insensitive
		{"memo.DocX", true},
		{"letter.doc", true},
		{"archive.zip", false},
		{"noext", false},
		{"weird.pdf.bak", false}, // only the final extension counts
	}
	for _, c := range cases {
		if got := matchesAnyExtension(c.name, exts); got != c.want {
			t.Errorf("matchesAnyExtension(%q) = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestPassesFilters(t *testing.T) {
	// No restrictions: everything passes.
	if !passesFilters("anything.xyz", nil, nil) {
		t.Error("expected all files to pass with no filters or extensions")
	}

	// Glob filter only (case-sensitive).
	if !passesFilters("main.go", []string{"*.go"}, nil) {
		t.Error("expected main.go to pass *.go filter")
	}
	if passesFilters("MAIN.GO", []string{"*.go"}, nil) {
		t.Error("expected MAIN.GO to fail case-sensitive *.go filter")
	}

	// Extensions only (case-insensitive).
	if !passesFilters("MAIN.GO", nil, []string{".go"}) {
		t.Error("expected MAIN.GO to pass case-insensitive .go extension")
	}

	// Combined: matches via either path.
	if !passesFilters("notes.txt", []string{"*.go"}, []string{".txt"}) {
		t.Error("expected notes.txt to pass via extension when glob misses")
	}
	if passesFilters("notes.md", []string{"*.go"}, []string{".txt"}) {
		t.Error("expected notes.md to fail when it matches neither")
	}
}

func TestBuildEntriesExtensionFilter(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"report.PDF", "memo.docx", "photo.jpg"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	raw, err := osFS{}.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir error: %v", err)
	}
	entries := buildEntries(osFS{}, dir, raw, false, nil, []string{".pdf", ".docx"})

	names := map[string]bool{}
	for _, e := range entries {
		names[e.Name] = true
	}
	if !names["report.PDF"] {
		t.Error("expected report.PDF to be visible (case-insensitive .pdf)")
	}
	if !names["memo.docx"] {
		t.Error("expected memo.docx to be visible")
	}
	if names["photo.jpg"] {
		t.Error("expected photo.jpg to be filtered out")
	}
}
