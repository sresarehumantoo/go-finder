package finder_test

import (
	"path/filepath"
	"strings"
	"testing"

	finder "github.com/SREsAreHumanToo/go-finder"
	tea "github.com/charmbracelet/bubbletea"
)

// loadPreview builds a model, sizes the terminal, and loads the start dir.
func loadPreview(t *testing.T, opts finder.Options, w, h int) finder.Model {
	t.Helper()
	m := finder.NewModel(opts)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	m = updated.(finder.Model)
	updated, _ = m.Update(m.Init()())
	return updated.(finder.Model)
}

func previewOpts(dir string) finder.Options {
	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = dir
	opts.Preview = true
	return opts
}

func TestPreviewOffByDefault(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "note.txt", "SECRET_PREVIEW_BODY")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = dir
	m := loadPreview(t, opts, 120, 40)

	if strings.Contains(m.View(), "SECRET_PREVIEW_BODY") {
		t.Error("file contents should not appear when preview is disabled")
	}
}

func TestPreviewTextFile(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "note.txt", "SECRET_PREVIEW_BODY")

	m := loadPreview(t, previewOpts(dir), 120, 40)

	if !strings.Contains(m.View(), "SECRET_PREVIEW_BODY") {
		t.Errorf("expected file head in preview pane:\n%s", m.View())
	}
}

func TestPreviewUpdatesOnCursorMove(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "a.txt", "AAA_BODY")
	createFile(t, dir, "b.txt", "BBB_BODY")

	m := loadPreview(t, previewOpts(dir), 120, 40)
	if !strings.Contains(m.View(), "AAA_BODY") {
		t.Fatalf("expected first file's preview:\n%s", m.View())
	}

	// Move down to b.txt.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(finder.Model)
	view := m.View()
	if !strings.Contains(view, "BBB_BODY") {
		t.Errorf("expected preview to update to second file:\n%s", view)
	}
	if strings.Contains(view, "AAA_BODY") {
		t.Error("expected previous file's preview to be replaced")
	}
}

func TestPreviewDirectoryListing(t *testing.T) {
	dir := t.TempDir()
	createDir(t, dir, "sub")
	createFile(t, filepath.Join(dir, "sub"), "CHILD_ENTRY.txt", "x")

	m := loadPreview(t, previewOpts(dir), 120, 40)

	// Cursor starts on "sub/" (directories sort first); its preview lists
	// the child.
	if !strings.Contains(m.View(), "CHILD_ENTRY.txt") {
		t.Errorf("expected directory listing in preview:\n%s", m.View())
	}
}

func TestPreviewBinaryFile(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "bin.dat", "abc\x00\x01\x02def")

	m := loadPreview(t, previewOpts(dir), 120, 40)

	if !strings.Contains(m.View(), "(binary file)") {
		t.Errorf("expected binary-file notice in preview:\n%s", m.View())
	}
}

func TestPreviewCustomFunc(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "note.txt", "ignored")

	opts := previewOpts(dir)
	opts.PreviewFunc = func(e finder.FileEntry, _, _ int) string {
		return "CUSTOM::" + e.Name
	}
	m := loadPreview(t, opts, 120, 40)

	if !strings.Contains(m.View(), "CUSTOM::note.txt") {
		t.Errorf("expected custom preview output:\n%s", m.View())
	}
}

func TestPreviewHiddenOnNarrowTerminal(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "note.txt", "SECRET_PREVIEW_BODY")

	// Width below previewMinTotalWidth (80) — pane must be suppressed.
	m := loadPreview(t, previewOpts(dir), 50, 40)

	view := m.View()
	if strings.Contains(view, "SECRET_PREVIEW_BODY") {
		t.Error("preview should be hidden on a narrow terminal")
	}
	if !strings.Contains(view, "note.txt") {
		t.Errorf("file list should still render on a narrow terminal:\n%s", view)
	}
}
