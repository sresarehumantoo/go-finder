package finder_test

import (
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	finder "github.com/SREsAreHumanToo/go-finder"
	tea "github.com/charmbracelet/bubbletea"
)

func sampleFS() fstest.MapFS {
	return fstest.MapFS{
		"alpha.go":     {Data: []byte("package main")},
		"beta.txt":     {Data: []byte("hello")},
		"sub/inner.go": {Data: []byte("inner")},
	}
}

// newFSModel builds a picker over a custom filesystem and loads the root.
func newFSModel(t *testing.T, fsys fs.FS, mode finder.Mode, interactive bool) finder.Model {
	t.Helper()
	opts := finder.DefaultOptions()
	opts.Mode = mode
	opts.Interactive = interactive
	finder.WithFS(fsys)(&opts)
	m := finder.NewModel(opts)
	updated, _ := m.Update(m.Init()())
	return updated.(finder.Model)
}

func send(t *testing.T, m finder.Model, msg tea.Msg) finder.Model {
	t.Helper()
	updated, cmd := m.Update(msg)
	m = updated.(finder.Model)
	if cmd != nil {
		if next := cmd(); next != nil {
			updated, _ = m.Update(next)
			m = updated.(finder.Model)
		}
	}
	return m
}

func TestFSNavigationAndSelection(t *testing.T) {
	m := newFSModel(t, sampleFS(), finder.ModeFile, false)

	view := m.View()
	for _, want := range []string{"alpha.go", "beta.txt", "sub/"} {
		if !strings.Contains(view, want) {
			t.Errorf("root view missing %q:\n%s", want, view)
		}
	}

	// "sub" sorts first (directories before files); navigate into it.
	m = send(t, m, tea.KeyMsg{Type: tea.KeyRight})
	view = m.View()
	if !strings.Contains(view, "inner.go") {
		t.Errorf("expected inner.go after entering sub:\n%s", view)
	}
	if !strings.Contains(view, "/sub") {
		t.Errorf("expected breadcrumb '/sub' after entering sub:\n%s", view)
	}

	// Select the file.
	m = send(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if got := m.SelectedPath(); got != "sub/inner.go" {
		t.Errorf("expected selection sub/inner.go, got %q", got)
	}
}

func TestFSBackNavigation(t *testing.T) {
	m := newFSModel(t, sampleFS(), finder.ModeFile, false)

	m = send(t, m, tea.KeyMsg{Type: tea.KeyRight}) // into sub
	if !strings.Contains(m.View(), "/sub") {
		t.Fatal("expected to be inside sub")
	}

	m = send(t, m, tea.KeyMsg{Type: tea.KeyEscape}) // back to root
	view := m.View()
	if !strings.Contains(view, "alpha.go") || !strings.Contains(view, "sub/") {
		t.Errorf("expected root contents after going back:\n%s", view)
	}
}

func TestFSReadOnlyInteractiveDegrades(t *testing.T) {
	m := newFSModel(t, sampleFS(), finder.ModeFile, true)

	// 'n' would normally open the new-file prompt; on a read-only FS it must
	// not, and should report the read-only status instead.
	m = send(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	view := m.View()
	if strings.Contains(view, "New file:") {
		t.Error("read-only FS should not open the new-file prompt")
	}
	if !strings.Contains(strings.ToLower(view), "read-only") {
		t.Errorf("expected read-only status message:\n%s", view)
	}
}

func TestFSFuzzySearch(t *testing.T) {
	m := newFSModel(t, sampleFS(), finder.ModeFile, false)

	// "aph" is a fuzzy subsequence of alpha.go but not beta.txt.
	m = typeSearch(t, m, "aph")
	view := m.View()
	if !strings.Contains(view, "alpha.go") {
		t.Error("expected alpha.go to fuzzy-match 'aph' on a custom FS")
	}
	if strings.Contains(view, "beta.txt") {
		t.Error("expected beta.txt to be filtered out")
	}
}
