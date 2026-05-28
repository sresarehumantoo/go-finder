package finder_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	finder "github.com/SREsAreHumanToo/go-finder"
	tea "github.com/charmbracelet/bubbletea"
)

func setupSearchModel(t *testing.T, dir string) finder.Model {
	t.Helper()
	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = dir
	m := finder.NewModel(opts)

	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	return updated.(finder.Model)
}

func TestSearchFiltersEntries(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "alpha.go", "")
	createFile(t, dir, "beta.go", "")
	createFile(t, dir, "gamma.txt", "")

	m := setupSearchModel(t, dir)

	// Press '/' to start search.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(finder.Model)

	// Type "alpha".
	for _, r := range "alpha" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(finder.Model)
	}

	view := m.View()
	if !strings.Contains(view, "alpha.go") {
		t.Error("expected alpha.go to be visible during search")
	}
	if strings.Contains(view, "beta.go") {
		t.Error("expected beta.go to be hidden during search")
	}
	if strings.Contains(view, "gamma.txt") {
		t.Error("expected gamma.txt to be hidden during search")
	}
}

func TestSearchEscRestoresAll(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "alpha.go", "")
	createFile(t, dir, "beta.go", "")

	m := setupSearchModel(t, dir)

	// Start search and type something.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(finder.Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(finder.Model)

	// Press esc to cancel search.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = updated.(finder.Model)

	// All entries should be restored.
	view := m.View()
	if !strings.Contains(view, "alpha.go") {
		t.Error("expected alpha.go after cancelling search")
	}
	if !strings.Contains(view, "beta.go") {
		t.Error("expected beta.go after cancelling search")
	}
}

func TestSearchEnterAcceptsFiltered(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "alpha.go", "")
	createFile(t, dir, "beta.go", "")

	m := setupSearchModel(t, dir)

	// Search for "beta".
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(finder.Model)
	for _, r := range "beta" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(finder.Model)
	}

	// Press enter to accept filtered results.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(finder.Model)

	// Should only show beta.go.
	view := m.View()
	if !strings.Contains(view, "beta.go") {
		t.Error("expected beta.go after accepting search")
	}
	if strings.Contains(view, "alpha.go") {
		t.Error("expected alpha.go to remain hidden after accepting search")
	}
}

func TestSearchBackspaceWidensResults(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "alpha.go", "")
	createFile(t, dir, "alto.go", "")
	createFile(t, dir, "beta.go", "")

	m := setupSearchModel(t, dir)

	// Search for "alph" — only alpha.go matches.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(finder.Model)
	for _, r := range "alph" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(finder.Model)
	}

	view := m.View()
	if strings.Contains(view, "alto.go") {
		t.Error("alto.go should be hidden when searching 'alph'")
	}

	// Backspace twice to "al" — both alpha.go and alto.go contain "al".
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = updated.(finder.Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = updated.(finder.Model)

	view = m.View()
	if !strings.Contains(view, "alpha.go") {
		t.Error("expected alpha.go with search 'al'")
	}
	if !strings.Contains(view, "alto.go") {
		t.Error("expected alto.go with search 'al'")
	}
	if strings.Contains(view, "beta.go") {
		t.Error("expected beta.go to remain hidden with search 'al'")
	}
}

func TestSearchCaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "README.md", "")
	createFile(t, dir, "readme.txt", "")
	createFile(t, dir, "other.go", "")

	m := setupSearchModel(t, dir)

	// Search for "readme" (lowercase) should match both.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(finder.Model)
	for _, r := range "readme" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(finder.Model)
	}

	view := m.View()
	if !strings.Contains(view, "README.md") {
		t.Error("expected README.md to match case-insensitive search")
	}
	if !strings.Contains(view, "readme.txt") {
		t.Error("expected readme.txt to match case-insensitive search")
	}
	if strings.Contains(view, "other.go") {
		t.Error("expected other.go to be hidden")
	}
}

func TestSearchEmptyTermShowsAll(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "alpha.go", "")
	createFile(t, dir, "beta.go", "")

	m := setupSearchModel(t, dir)

	// Start search.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(finder.Model)

	// Type a character then backspace to empty.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m = updated.(finder.Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = updated.(finder.Model)

	// All entries should be visible with empty search term.
	view := m.View()
	if !strings.Contains(view, "alpha.go") || !strings.Contains(view, "beta.go") {
		t.Error("expected all entries visible with empty search term")
	}
}

// typeSearch starts a search and types the given term into the model.
func typeSearch(t *testing.T, m finder.Model, term string) finder.Model {
	t.Helper()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = updated.(finder.Model)
	for _, r := range term {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(finder.Model)
	}
	return m
}

func TestFuzzySearchRanksContiguousHigher(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "cfg.go", "")    // contiguous prefix match for "cfg"
	createFile(t, dir, "config.go", "") // scattered subsequence c..f..g
	createFile(t, dir, "readme.md", "") // no c-f-g subsequence

	m := setupSearchModel(t, dir)
	m = typeSearch(t, m, "cfg")

	view := m.View()
	cfgIdx := strings.Index(view, "cfg.go")
	configIdx := strings.Index(view, "config.go")

	if cfgIdx == -1 || configIdx == -1 {
		t.Fatalf("expected both cfg.go and config.go to match 'cfg'; view:\n%s", view)
	}
	if cfgIdx > configIdx {
		t.Errorf("expected cfg.go (tighter match) to rank above config.go")
	}
	if strings.Contains(view, "readme.md") {
		t.Error("expected readme.md to be excluded from 'cfg' search")
	}
}

func TestFuzzySearchMatchesSubsequence(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "alpha.go", "") // matches subsequence a-g-o
	createFile(t, dir, "beta.txt", "") // does not

	m := setupSearchModel(t, dir)
	m = typeSearch(t, m, "ago")

	view := m.View()
	if !strings.Contains(view, "alpha.go") {
		t.Error("expected alpha.go to match subsequence 'ago' with fuzzy search")
	}
	if strings.Contains(view, "beta.txt") {
		t.Error("expected beta.txt to be excluded from 'ago' search")
	}
}

func TestSubstringSearchOptOut(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "alpha.go", "")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = dir
	finder.WithFuzzySearch(false)(&opts)
	m := finder.NewModel(opts)
	cmd := m.Init()
	updated, _ := m.Update(cmd())
	m = updated.(finder.Model)

	// With substring matching, "ago" is NOT a substring of "alpha.go".
	m = typeSearch(t, m, "ago")
	view := m.View()
	if strings.Contains(view, "alpha.go") {
		t.Error("substring opt-out should not match 'ago' against alpha.go")
	}
}

func TestCreateFileTrailingSlashCreatesDir(t *testing.T) {
	dir := t.TempDir()
	m := setupInteractiveModel(t, dir)

	// Press 'n' to start new file prompt.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(finder.Model)

	// Type "mydir/" — trailing slash means create directory.
	for _, r := range "mydir/" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(finder.Model)
	}

	// Press enter.
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("expected readDir command after creating")
	}

	// Verify a directory (not a file) was created.
	info, err := os.Stat(filepath.Join(dir, "mydir"))
	if err != nil {
		t.Fatalf("expected mydir to exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected mydir to be a directory, but it's a file")
	}
}
