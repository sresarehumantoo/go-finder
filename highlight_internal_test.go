package finder

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// TestFilterFuzzyRecordsMatchOffsets verifies that fuzzy filtering records, for
// each matched entry, the byte offsets of the characters that matched the query.
func TestFilterFuzzyRecordsMatchOffsets(t *testing.T) {
	m := &Model{
		options:    Options{FuzzySearch: true},
		searchTerm: "ago",
		allEntries: []FileEntry{
			{Name: "alpha.go", Path: "alpha.go"},
			{Name: "beta.txt", Path: "beta.txt"},
		},
	}

	m.filterEntries()

	if len(m.entries) != 1 || m.entries[0].Name != "alpha.go" {
		t.Fatalf("expected only alpha.go to match, got %+v", m.entries)
	}

	idx, ok := m.matchIdx["alpha.go"]
	if !ok || len(idx) == 0 {
		t.Fatalf("expected match offsets for alpha.go, got %v (ok=%v)", idx, ok)
	}

	// The matched offsets, read off the name in order, must spell the query.
	const name = "alpha.go"
	var got []byte
	for _, i := range idx {
		if i < 0 || i >= len(name) {
			t.Fatalf("offset %d out of range for %q", i, name)
		}
		got = append(got, name[i])
	}
	if string(got) != "ago" {
		t.Errorf("matched characters = %q, want %q (offsets %v)", got, "ago", idx)
	}

	if _, ok := m.matchIdx["beta.txt"]; ok {
		t.Error("beta.txt should not have match offsets")
	}
}

// TestFilterEntriesClearsMatchIdx verifies the match map is reset for the
// substring path and the empty-term path.
func TestFilterEntriesClearsMatchIdx(t *testing.T) {
	m := &Model{
		options:    Options{FuzzySearch: false},
		searchTerm: "al",
		matchIdx:   map[string][]int{"stale": {0}},
		allEntries: []FileEntry{{Name: "alpha.go", Path: "alpha.go"}},
	}
	m.filterEntries()
	if m.matchIdx != nil {
		t.Errorf("expected matchIdx cleared in substring mode, got %v", m.matchIdx)
	}
}

// TestHighlightNamePreservesText guards the offset math: in the no-color test
// environment the styles render plain, so the output must equal the input with
// no dropped, duplicated, or reordered characters.
func TestHighlightNamePreservesText(t *testing.T) {
	base := lipgloss.NewStyle()
	match := lipgloss.NewStyle().Bold(true)

	cases := []struct {
		name    string
		matched []int
	}{
		{"alpha.go", []int{0, 6, 7}},
		{"x.go", nil},
		{"café.txt", []int{0}}, // multibyte: 'é' must not be split
		{"README.md", []int{2, 3}},
	}
	for _, tc := range cases {
		if got := highlightName(tc.name, base, match, tc.matched); got != tc.name {
			t.Errorf("highlightName(%q, %v) = %q, want %q", tc.name, tc.matched, got, tc.name)
		}
	}
}
