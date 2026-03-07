package finder_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	finder "github.com/SREsAreHumanToo/go-finder"
)

func TestExpandSymlinksStartDir(t *testing.T) {
	// Create a real directory and a symlink pointing to it.
	realDir := t.TempDir()
	createFile(t, realDir, "real.txt", "content")

	linkParent := t.TempDir()
	linkPath := filepath.Join(linkParent, "mylink")
	if err := os.Symlink(realDir, linkPath); err != nil {
		t.Skip("symlinks not supported:", err)
	}

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = linkPath
	opts.ExpandSymlinks = true
	m := finder.NewModel(opts)

	// Load entries.
	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// The view should show the real path, not the symlink path.
	view := m.View()
	if filepath.Clean(realDir) == filepath.Clean(linkPath) {
		t.Skip("real dir and link path are the same")
	}

	// The displayed path should be the resolved real directory.
	realAbs, _ := filepath.Abs(realDir)
	if !containsPath(view, realAbs) {
		t.Errorf("expected resolved path %s in view, got:\n%s", realAbs, view)
	}
}

func TestExpandSymlinksBackNavigation(t *testing.T) {
	// Create /tmp/real/child/ and symlink /tmp/links/mylink -> /tmp/real/child
	realBase := t.TempDir()
	realChild := filepath.Join(realBase, "child")
	if err := os.Mkdir(realChild, 0755); err != nil {
		t.Fatal(err)
	}
	createFile(t, realChild, "file.txt", "data")

	linkParent := t.TempDir()
	linkPath := filepath.Join(linkParent, "mylink")
	if err := os.Symlink(realChild, linkPath); err != nil {
		t.Skip("symlinks not supported:", err)
	}

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = linkPath
	opts.ExpandSymlinks = true
	m := finder.NewModel(opts)

	// Load entries.
	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Press esc to go back — should go to realBase, NOT linkParent.
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = updated.(finder.Model)

	if cmd == nil {
		t.Fatal("expected readDir command after backing out")
	}

	// Complete the read.
	msg = cmd()
	updated, _ = m.Update(msg)
	m = updated.(finder.Model)

	// The view should show realBase as the current path.
	view := m.View()
	realBaseAbs, _ := filepath.Abs(realBase)
	if !containsPath(view, realBaseAbs) {
		t.Errorf("expected real parent %s in view after esc, got:\n%s", realBaseAbs, view)
	}

	// Should NOT contain the link parent path.
	linkParentAbs, _ := filepath.Abs(linkParent)
	if containsPath(view, linkParentAbs) {
		t.Errorf("should not show symlink parent %s, got:\n%s", linkParentAbs, view)
	}
}

func TestNoExpandSymlinksDefault(t *testing.T) {
	// Without ExpandSymlinks, the symlink path should be preserved.
	realDir := t.TempDir()
	createFile(t, realDir, "real.txt", "content")

	linkParent := t.TempDir()
	linkPath := filepath.Join(linkParent, "mylink")
	if err := os.Symlink(realDir, linkPath); err != nil {
		t.Skip("symlinks not supported:", err)
	}

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = linkPath
	// ExpandSymlinks = false (default)
	m := finder.NewModel(opts)

	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// The view should show the symlink path, not the resolved path.
	view := m.View()
	linkAbs, _ := filepath.Abs(linkPath)
	if !containsPath(view, linkAbs) {
		t.Errorf("expected symlink path %s in view, got:\n%s", linkAbs, view)
	}
}

func TestResolvePath(t *testing.T) {
	realDir := t.TempDir()
	linkParent := t.TempDir()
	linkPath := filepath.Join(linkParent, "testlink")
	if err := os.Symlink(realDir, linkPath); err != nil {
		t.Skip("symlinks not supported:", err)
	}

	resolved := finder.ResolvePath(linkPath)
	realAbs, _ := filepath.Abs(realDir)
	if resolved != realAbs {
		t.Errorf("ResolvePath(%s) = %s, want %s", linkPath, resolved, realAbs)
	}
}

// containsPath checks if the view string contains the given path.
func containsPath(view, path string) bool {
	return strings.Contains(view, filepath.Clean(path))
}
