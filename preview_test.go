package finder_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	finder "github.com/SREsAreHumanToo/go-finder"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

// TestPreviewStableHeight guards against the layout jumping as the selection
// moves between short and long previews: the rendered view must keep a constant
// height.
func TestPreviewStableHeight(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "a_long.txt", strings.Repeat("x\n", 200))
	createFile(t, dir, "b_short.txt", "one line")

	m := loadPreview(t, previewOpts(dir), 100, 24)
	tall := strings.Count(m.View(), "\n") // cursor on a_long.txt

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(finder.Model)
	short := strings.Count(m.View(), "\n") // cursor on b_short.txt

	if tall != short {
		t.Errorf("view height changed with selection (%d vs %d lines) — preview pane is jumping", tall, short)
	}
}

// TestPreviewFitsTerminalWhileScrolling reproduces the "jumping" report: with a
// scrollable list and a narrow terminal (where the help bar wraps), holding the
// down arrow must not produce frames taller than the terminal, and the frame
// height must stay constant.
func TestPreviewFitsTerminalWhileScrolling(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 40; i++ {
		// Vary content length so previews differ in height.
		body := strings.Repeat("x\n", (i*9)%150)
		createFile(t, dir, fmt.Sprintf("file_%02d.txt", i), body)
	}

	const height = 24
	m := loadPreview(t, previewOpts(dir), 60, height) // narrow → help bar wraps

	first := strings.Count(m.View(), "\n") + 1
	for i := 0; i < 40; i++ {
		n := strings.Count(m.View(), "\n") + 1
		if n > height {
			t.Fatalf("frame %d is %d lines, exceeds terminal height %d (would scroll/jump)", i, n, height)
		}
		if n != first {
			t.Fatalf("frame %d height %d != initial %d (layout jumping)", i, n, first)
		}
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updated.(finder.Model)
	}
}

// TestViewNeverWrapsOrOverflows guards the terminal-corruption fix: no rendered
// line may exceed the terminal width (which the terminal would wrap into extra
// rows, e.g. a long title), and the frame must not exceed the terminal height —
// at any size, including very small ones.
func TestViewNeverWrapsOrOverflows(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 40; i++ {
		createFile(t, dir, fmt.Sprintf("file_%02d.txt", i), strings.Repeat("x\n", (i*9)%150))
	}

	sizes := []struct{ w, h int }{{100, 24}, {60, 24}, {40, 20}, {30, 15}}
	for _, s := range sizes {
		opts := previewOpts(dir)
		opts.Title = "Pick a file (→/l to enter dirs, / to search)" // long, multibyte
		m := finder.NewModel(opts)
		u, _ := m.Update(tea.WindowSizeMsg{Width: s.w, Height: s.h})
		m = u.(finder.Model)
		u, _ = m.Update(m.Init()())
		m = u.(finder.Model)

		for i := 0; i < 40; i++ {
			lines := strings.Split(m.View(), "\n")
			if len(lines) > s.h {
				t.Errorf("%dx%d frame %d: %d rows exceeds height %d", s.w, s.h, i, len(lines), s.h)
			}
			for _, ln := range lines {
				if w := lipgloss.Width(ln); w > s.w {
					t.Errorf("%dx%d frame %d: line width %d exceeds %d: %q", s.w, s.h, i, w, s.w, ln)
				}
			}
			u, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
			m = u.(finder.Model)
		}
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
