package finder_test

import (
	"runtime"
	"strings"
	"testing"

	finder "github.com/rummage-dev/rummage"
	tea "github.com/charmbracelet/bubbletea"
)

// renderDir builds a model over dir, runs the initial directory read, and
// returns the rendered view — the exact bytes that would be written to the
// terminal.
func renderDir(t *testing.T, opts finder.Options) string {
	t.Helper()
	m := finder.NewModel(opts)
	msg := m.Init()()
	updated, _ := m.Update(msg)
	updated, _ = updated.(finder.Model).Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return updated.(finder.Model).View()
}

// TestViewSanitizesMaliciousFilename verifies that a filename carrying terminal
// escape sequences cannot reach the rendered output. Without sanitization the
// raw ESC bytes would be written straight to the user's terminal, allowing
// window-title rewrites, cursor manipulation, or OSC 52 clipboard hijacking
// when browsing an attacker-controlled directory.
func TestViewSanitizesMaliciousFilename(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows forbids control characters in filenames")
	}
	dir := t.TempDir()
	// A name with an SGR sequence and an OSC set-window-title sequence.
	createFile(t, dir, "evil\x1b[31m\x1b]0;pwned\x07.txt", "x")

	opts := finder.DefaultOptions()
	opts.StartDir = dir
	out := renderDir(t, opts)

	// Legitimate lipgloss styling emits SGR (ESC [ … m) sequences, so the test
	// targets the markers an attacker needs that styling never produces: OSC
	// (ESC ]) and BEL (the OSC/title terminator).
	if strings.Contains(out, "\x1b]") || strings.Contains(out, "\x07") {
		t.Fatalf("view leaked a raw escape/control sequence from a filename:\n%q", out)
	}
}

// TestViewSanitizesPreviewContent verifies that escape sequences inside a
// previewed file's *contents* are neutralized before display. This is
// cross-platform because file contents (unlike filenames) may contain control
// bytes on every OS.
func TestViewSanitizesPreviewContent(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "notes.txt", "hello\x1b]0;pwned\x07\x1b[2Jworld")

	opts := finder.DefaultOptions()
	opts.StartDir = dir
	opts.Preview = true
	out := renderDir(t, opts)

	if strings.Contains(out, "\x1b]") || strings.Contains(out, "\x07") {
		t.Fatalf("view leaked a raw OSC/BEL sequence from preview content:\n%q", out)
	}
	// The benign text around the escapes should still be previewed.
	if !strings.Contains(out, "hello") || !strings.Contains(out, "world") {
		t.Errorf("expected preview to keep the readable text, got:\n%q", out)
	}
}
