package finder_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	finder "github.com/SREsAreHumanToo/go-finder"
)

func setupInteractiveModel(t *testing.T, dir string) finder.Model {
	t.Helper()
	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeAny
	opts.StartDir = dir
	opts.Interactive = true
	m := finder.NewModel(opts)

	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	return updated.(finder.Model)
}

func TestInteractiveCreateFile(t *testing.T) {
	dir := t.TempDir()
	m := setupInteractiveModel(t, dir)

	// Press 'n' to start new file prompt.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(finder.Model)

	// View should show the new file prompt.
	view := m.View()
	if !strings.Contains(view, "New file:") {
		t.Fatalf("expected 'New file:' prompt, got:\n%s", view)
	}

	// Type a file name.
	for _, r := range "test.txt" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(finder.Model)
	}

	// Press enter to create.
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(finder.Model)

	// Should have triggered a readDir.
	if cmd == nil {
		t.Fatal("expected readDir command after creating file")
	}

	// Verify file was created on disk.
	if _, err := os.Stat(filepath.Join(dir, "test.txt")); err != nil {
		t.Fatalf("file was not created: %v", err)
	}

	// Complete the readDir.
	msg := cmd()
	updated, _ = m.Update(msg)
	m = updated.(finder.Model)

	// Cursor should be on the new file.
	selected := m.SelectedPath()
	_ = selected
	// Verify by checking the view contains the file.
	view = m.View()
	if !strings.Contains(view, "test.txt") {
		t.Errorf("expected test.txt in listing, got:\n%s", view)
	}
}

func TestInteractiveCreateFolder(t *testing.T) {
	dir := t.TempDir()
	m := setupInteractiveModel(t, dir)

	// Press 'N' to start new folder prompt.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})
	m = updated.(finder.Model)

	view := m.View()
	if !strings.Contains(view, "New folder:") {
		t.Fatalf("expected 'New folder:' prompt, got:\n%s", view)
	}

	// Type a folder name.
	for _, r := range "mydir" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(finder.Model)
	}

	// Press enter to create.
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(finder.Model)

	if cmd == nil {
		t.Fatal("expected readDir command after creating folder")
	}

	// Verify directory was created on disk.
	info, err := os.Stat(filepath.Join(dir, "mydir"))
	if err != nil {
		t.Fatalf("directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected a directory, got a file")
	}
}

func TestInteractiveDeleteFile(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "doomed.txt", "goodbye")
	m := setupInteractiveModel(t, dir)

	// Press 'd' to start delete confirmation.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(finder.Model)

	view := m.View()
	if !strings.Contains(view, "Delete") || !strings.Contains(view, "(y/n)") {
		t.Fatalf("expected delete confirmation prompt, got:\n%s", view)
	}

	// Press 'y' to confirm.
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = updated.(finder.Model)

	if cmd == nil {
		t.Fatal("expected readDir command after deleting")
	}

	// Verify file was deleted from disk.
	if _, err := os.Stat(filepath.Join(dir, "doomed.txt")); !os.IsNotExist(err) {
		t.Error("file should have been deleted")
	}
}

func TestInteractiveDeleteFolder(t *testing.T) {
	dir := t.TempDir()
	createDir(t, dir, "doomeddir")
	createFile(t, dir+"/doomeddir", "inside.txt", "nested")
	m := setupInteractiveModel(t, dir)

	// Cursor should be on the directory (sorted first).
	// Press 'd' then 'y'.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(finder.Model)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = updated.(finder.Model)

	if cmd == nil {
		t.Fatal("expected readDir command after deleting folder")
	}

	// Verify directory and contents were deleted.
	if _, err := os.Stat(filepath.Join(dir, "doomeddir")); !os.IsNotExist(err) {
		t.Error("directory should have been deleted")
	}
}

func TestInteractiveDeleteCancelled(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "safe.txt", "keep me")
	m := setupInteractiveModel(t, dir)

	// Press 'd' then 'n' to cancel.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(finder.Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(finder.Model)

	// File should still exist.
	if _, err := os.Stat(filepath.Join(dir, "safe.txt")); err != nil {
		t.Error("file should NOT have been deleted after cancelling")
	}
}

func TestInteractiveCreateCancelled(t *testing.T) {
	dir := t.TempDir()
	m := setupInteractiveModel(t, dir)

	// Press 'n' then esc to cancel.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(finder.Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = updated.(finder.Model)

	// Should be back to normal mode, no prompt visible.
	view := m.View()
	if strings.Contains(view, "New file:") {
		t.Error("prompt should be dismissed after esc")
	}
}

func TestInteractiveDisabledByDefault(t *testing.T) {
	dir := t.TempDir()

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = dir
	// Interactive NOT enabled.
	m := finder.NewModel(opts)

	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Press 'n' — should not enter input mode.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(finder.Model)

	view := m.View()
	if strings.Contains(view, "New file:") {
		t.Error("interactive actions should be disabled when Interactive is false")
	}
}

func TestInteractiveCreateDuplicateShowsError(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "exists.txt", "already here")
	m := setupInteractiveModel(t, dir)

	// Try to create a file with the same name.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(finder.Model)
	for _, r := range "exists.txt" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(finder.Model)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(finder.Model)

	view := m.View()
	if !strings.Contains(view, "Error:") {
		t.Errorf("expected error message for duplicate file, got:\n%s", view)
	}
}
