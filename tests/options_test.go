package tests

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	finder "github.com/SREsAreHumanToo/go-finder/src/finder"
)

func TestDefaultOptions(t *testing.T) {
	opts := finder.DefaultOptions()

	if opts.StartDir == "" {
		t.Error("expected StartDir to be set")
	}
	if opts.ShowHidden {
		t.Error("expected ShowHidden to default to false")
	}
	if opts.Mode != finder.ModeFile {
		t.Errorf("expected Mode to be ModeFile, got %d", opts.Mode)
	}
	if opts.Title == "" {
		t.Error("expected Title to be set")
	}
}

func TestWithStartDir(t *testing.T) {
	opts := finder.DefaultOptions()
	finder.WithStartDir("/tmp")(&opts)

	if opts.StartDir != "/tmp" {
		t.Errorf("expected StartDir /tmp, got %s", opts.StartDir)
	}
}

func TestWithFilter(t *testing.T) {
	opts := finder.DefaultOptions()
	finder.WithFilter("*.go", "*.txt")(&opts)

	if len(opts.Filters) != 2 {
		t.Fatalf("expected 2 filters, got %d", len(opts.Filters))
	}
	if opts.Filters[0] != "*.go" || opts.Filters[1] != "*.txt" {
		t.Errorf("unexpected filters: %v", opts.Filters)
	}
}

func TestWithHidden(t *testing.T) {
	opts := finder.DefaultOptions()
	finder.WithHidden(true)(&opts)

	if !opts.ShowHidden {
		t.Error("expected ShowHidden to be true")
	}
}

func TestWithTitle(t *testing.T) {
	opts := finder.DefaultOptions()
	finder.WithTitle("Pick something")(&opts)

	if opts.Title != "Pick something" {
		t.Errorf("expected title 'Pick something', got %q", opts.Title)
	}
}

func TestWithHeight(t *testing.T) {
	opts := finder.DefaultOptions()
	finder.WithHeight(20)(&opts)

	if opts.Height != 20 {
		t.Errorf("expected height 20, got %d", opts.Height)
	}
}

func TestWithMode(t *testing.T) {
	opts := finder.DefaultOptions()
	finder.WithMode(finder.ModeFolder)(&opts)

	if opts.Mode != finder.ModeFolder {
		t.Errorf("expected ModeFolder, got %d", opts.Mode)
	}
}

func TestWithKeyMap(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "file.txt", "data")

	// Custom keymap: use 'x' for cancel instead of 'q'.
	km := finder.DefaultKeyMap()
	km.Cancel = key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "exit"),
	)

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = dir
	opts.KeyMap = &km
	m := finder.NewModel(opts)

	// Load entries.
	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// 'q' should NOT quit (we rebound cancel to 'x').
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updated.(finder.Model)

	// Should still be running (view should have content).
	view := m.View()
	if view == "" {
		t.Error("'q' should not quit when cancel is rebound to 'x'")
	}

	// Verify the help bar shows the custom key.
	if !strings.Contains(view, "x") {
		t.Error("expected custom key 'x' in help bar")
	}
}

func TestWithKeyMapCustomSelectKey(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "target.txt", "data")

	// Custom keymap: use 'f' for select instead of 'enter'.
	km := finder.DefaultKeyMap()
	km.Select = key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "pick"),
	)

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = dir
	opts.KeyMap = &km
	m := finder.NewModel(opts)

	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// 'f' should select the file.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m = updated.(finder.Model)

	selected := m.SelectedPath()
	if selected == "" {
		t.Error("custom select key 'f' should have selected the file")
	}
}
