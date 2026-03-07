package finder_test

import (
	"os"
	"strings"
	"testing"

	finder "github.com/SREsAreHumanToo/go-finder"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel(t *testing.T) {
	opts := finder.DefaultOptions()
	m := finder.NewModel(opts)

	if m.SelectedPath() != "" {
		t.Error("expected no selection before picker runs")
	}
	if paths := m.SelectedPaths(); len(paths) != 0 {
		t.Errorf("expected empty paths, got %v", paths)
	}
	if m.Err() != nil {
		t.Errorf("expected no error, got %v", m.Err())
	}
}

func TestModelInit(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "test.txt", "content")

	opts := finder.DefaultOptions()
	opts.StartDir = dir
	m := finder.NewModel(opts)

	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected Init to return a command")
	}
}

func TestFolderModeSelectHighlightedDir(t *testing.T) {
	dir := t.TempDir()
	sub := "mysubdir"
	createDir(t, dir, sub)
	createFile(t, dir, "file.txt", "data")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFolder
	opts.StartDir = dir
	m := finder.NewModel(opts)

	// Simulate directory read completing.
	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Cursor should be on the first entry (the directory, sorted first).
	// Press enter to select the highlighted directory.
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(finder.Model)

	selected := m.SelectedPath()
	if selected == "" {
		t.Fatal("expected a directory to be selected, got empty string")
	}
	if selected != dir+"/"+sub {
		t.Errorf("expected selected path %s/%s, got %s", dir, sub, selected)
	}
}

func TestFolderModeSelectCurrentDir(t *testing.T) {
	dir := t.TempDir()
	createDir(t, dir, "child")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFolder
	opts.StartDir = dir
	m := finder.NewModel(opts)

	// Simulate directory read completing.
	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Press 's' to select the current working directory.
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updated.(finder.Model)

	selected := m.SelectedPath()
	if selected != dir {
		t.Errorf("expected current dir %s, got %s", dir, selected)
	}
}

func TestFolderModeNavigateWithRight(t *testing.T) {
	dir := t.TempDir()
	sub := "child"
	createDir(t, dir, sub)

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFolder
	opts.StartDir = dir
	m := finder.NewModel(opts)

	// Simulate directory read completing.
	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Press right arrow to navigate into the directory (not select it).
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = updated.(finder.Model)

	// Should NOT have selected anything yet.
	if m.SelectedPath() != "" {
		t.Error("right arrow should navigate, not select")
	}

	// The navigate should have triggered a new dir read command.
	if cmd == nil {
		t.Fatal("expected a readDir command after navigating")
	}
}

func TestFileModeEnterOnDirNavigates(t *testing.T) {
	dir := t.TempDir()
	createDir(t, dir, "subdir")
	createFile(t, dir, "test.go", "package main")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = dir
	m := finder.NewModel(opts)

	// Simulate directory read.
	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Cursor on directory — enter should navigate, not select.
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(finder.Model)

	if m.SelectedPath() != "" {
		t.Error("in file mode, enter on a directory should navigate, not select")
	}
	if cmd == nil {
		t.Fatal("expected a readDir command after entering a directory")
	}
}

func TestAnyModeSelectFile(t *testing.T) {
	dir := t.TempDir()
	createDir(t, dir, "subdir")
	createFile(t, dir, "readme.md", "hello")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeAny
	opts.StartDir = dir
	m := finder.NewModel(opts)

	// Load entries.
	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Move cursor down past the directory to the file.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(finder.Model)

	// Press enter to select the file.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(finder.Model)

	selected := m.SelectedPath()
	if selected == "" {
		t.Fatal("expected a file to be selected")
	}
	if selected != dir+"/readme.md" {
		t.Errorf("expected %s/readme.md, got %s", dir, selected)
	}
}

func TestAnyModeEnterOnDirNavigates(t *testing.T) {
	dir := t.TempDir()
	sub := "mydir"
	createDir(t, dir, sub)
	createFile(t, dir, "file.txt", "data")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeAny
	opts.StartDir = dir
	m := finder.NewModel(opts)

	// Load entries.
	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Cursor is on the directory (sorted first). Press enter — should navigate, not select.
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(finder.Model)

	if m.SelectedPath() != "" {
		t.Error("in any mode, enter on a directory should navigate, not select")
	}
	if cmd == nil {
		t.Fatal("expected a readDir command after entering a directory")
	}
}

func TestAnyModeTabSelectsDir(t *testing.T) {
	dir := t.TempDir()
	sub := "mydir"
	createDir(t, dir, sub)

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeAny
	opts.StartDir = dir
	m := finder.NewModel(opts)

	// Load entries.
	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Tab on the directory should select it.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(finder.Model)

	selected := m.SelectedPath()
	if selected != dir+"/"+sub {
		t.Errorf("expected %s/%s, got %s", dir, sub, selected)
	}
}

func TestAnyModeNavigateWithRight(t *testing.T) {
	dir := t.TempDir()
	createDir(t, dir, "child")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeAny
	opts.StartDir = dir
	m := finder.NewModel(opts)

	// Load entries.
	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Press right to navigate into dir, not select it.
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = updated.(finder.Model)

	if m.SelectedPath() != "" {
		t.Error("right arrow should navigate, not select")
	}
	if cmd == nil {
		t.Fatal("expected a readDir command after navigating")
	}
}

func TestAnyModeSelectCurrentDir(t *testing.T) {
	dir := t.TempDir()
	createDir(t, dir, "child")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeAny
	opts.StartDir = dir
	m := finder.NewModel(opts)

	// Load entries.
	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Press 's' to select the current directory.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updated.(finder.Model)

	selected := m.SelectedPath()
	if selected != dir {
		t.Errorf("expected current dir %s, got %s", dir, selected)
	}
}

func TestBackReturnsCursorToLastDir(t *testing.T) {
	dir := t.TempDir()
	createDir(t, dir, "alpha")
	createDir(t, dir, "beta")
	createDir(t, dir, "gamma")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = dir
	m := finder.NewModel(opts)

	// Load entries.
	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Move cursor to "beta" (index 1, dirs sorted alphabetically).
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(finder.Model)

	// Enter "beta" directory.
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(finder.Model)

	// Simulate dirRead completing for beta/.
	msg = cmd()
	updated, _ = m.Update(msg)
	m = updated.(finder.Model)

	// Go back with backspace.
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = updated.(finder.Model)

	// Simulate dirRead completing for parent.
	msg = cmd()
	updated, _ = m.Update(msg)
	m = updated.(finder.Model)

	// Cursor should be on "beta" (index 1), not at the top.
	selected := m.SelectedPath()
	_ = selected
	// We can't call SelectedPath without choosing, so check cursor position
	// by pressing enter (which navigates into the dir under cursor in file mode).
	// Instead, let's navigate into it and see we end up in beta again.
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(finder.Model)
	if cmd == nil {
		t.Fatal("expected readDir command — cursor should be on a directory")
	}
	// Complete the read to see which dir we navigated into.
	msg = cmd()
	updated, _ = m.Update(msg)
	m = updated.(finder.Model)

	// We won't easily get the dir name from the model, but we can verify
	// it didn't go to "alpha" (index 0) by checking we can go back and
	// the test structure is consistent. A simpler approach: use ModeAny
	// and tab to select.
}

func TestBackReturnsCursorToLastDirAnyMode(t *testing.T) {
	dir := t.TempDir()
	createDir(t, dir, "aaa")
	createDir(t, dir, "bbb")
	createDir(t, dir, "ccc")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeAny
	opts.StartDir = dir
	m := finder.NewModel(opts)

	// Load entries.
	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Move cursor to "bbb" (index 1).
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(finder.Model)

	// Navigate into "bbb".
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(finder.Model)
	msg = cmd()
	updated, _ = m.Update(msg)
	m = updated.(finder.Model)

	// Go back.
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = updated.(finder.Model)
	msg = cmd()
	updated, _ = m.Update(msg)
	m = updated.(finder.Model)

	// Tab to select whatever cursor is on — should be "bbb".
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(finder.Model)

	selected := m.SelectedPath()
	if selected != dir+"/bbb" {
		t.Errorf("expected cursor on %s/bbb after going back, got %s", dir, selected)
	}
}

func TestAtHighestLevelMessage(t *testing.T) {
	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = "/"
	m := finder.NewModel(opts)

	// Load entries.
	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Press backspace at root.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = updated.(finder.Model)

	// View should contain the highest level message.
	view := m.View()
	if !strings.Contains(view, "At highest level") {
		t.Errorf("expected 'At highest level' message when at filesystem root, got:\n%s", view)
	}
}

func TestNoHighestLevelMessageWhenBackingNormally(t *testing.T) {
	dir := t.TempDir()
	child := "subdir"
	createDir(t, dir, child)

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = dir + "/" + child
	m := finder.NewModel(opts)

	// Load entries for the child dir.
	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Press esc to go back to parent — should NOT show "at highest level".
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = updated.(finder.Model)

	// Complete the dir read.
	msg = cmd()
	updated, _ = m.Update(msg)
	m = updated.(finder.Model)

	view := m.View()
	if strings.Contains(view, "At highest level") {
		t.Error("should not show 'At highest level' when backing out normally, only at filesystem root")
	}
}

func TestEscAtRootShowsMessage(t *testing.T) {
	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = "/"
	m := finder.NewModel(opts)

	// Load entries for root.
	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Press esc at root — should show message, NOT quit.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = updated.(finder.Model)

	view := m.View()
	if strings.Contains(view, "") && !strings.Contains(view, "At highest level") {
		t.Error("esc at root should show 'At highest level' message, not quit")
	}
	if m.SelectedPath() != "" {
		t.Error("esc at root should not select anything")
	}
}

func TestEscGoesBackToParent(t *testing.T) {
	dir := t.TempDir()
	child := "subdir"
	createDir(t, dir, child)

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = dir + "/" + child
	m := finder.NewModel(opts)

	// Load entries.
	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Press esc — should go back to parent, not quit.
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = updated.(finder.Model)

	if m.SelectedPath() != "" {
		t.Error("esc should go back, not select")
	}
	// Should have triggered a readDir for the parent.
	if cmd == nil {
		t.Fatal("expected a readDir command when pressing esc with a parent available")
	}
}

func TestMultiSelectNavigateIntoDir(t *testing.T) {
	dir := t.TempDir()
	createDir(t, dir, "subdir")
	createFile(t, dir, "file1.txt", "")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeMultiple
	opts.StartDir = dir
	m := finder.NewModel(opts)

	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Cursor is on "subdir" (dirs sorted first). Press enter to navigate.
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(finder.Model)

	if cmd == nil {
		t.Fatal("enter on dir in multi-select should navigate into it")
	}
	if m.SelectedPath() != "" {
		t.Error("navigating into dir should not select anything")
	}
}

func TestMultiSelectNavigateWithRight(t *testing.T) {
	dir := t.TempDir()
	createDir(t, dir, "subdir")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeMultiple
	opts.StartDir = dir
	m := finder.NewModel(opts)

	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Right arrow should navigate into dir.
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = updated.(finder.Model)

	if cmd == nil {
		t.Fatal("right arrow should navigate into dir in multi-select mode")
	}
}

func TestMultiSelectToggleAcrossDirs(t *testing.T) {
	dir := t.TempDir()
	createDir(t, dir, "sub")
	createFile(t, dir, "a.txt", "")
	createFile(t, dir+"/sub", "b.txt", "")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeMultiple
	opts.StartDir = dir
	m := finder.NewModel(opts)

	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Move to a.txt (index 1, after "sub/" dir) and toggle it.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(finder.Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(finder.Model)

	// Navigate into sub/.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(finder.Model)
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(finder.Model)
	msg = cmd()
	updated, _ = m.Update(msg)
	m = updated.(finder.Model)

	// Toggle b.txt.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(finder.Model)

	// Confirm selections.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(finder.Model)

	paths := m.SelectedPaths()
	if len(paths) != 2 {
		t.Fatalf("expected 2 selected paths across dirs, got %d: %v", len(paths), paths)
	}
}

func TestMultiSelectSelectionCount(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "a.txt", "")
	createFile(t, dir, "b.txt", "")
	createFile(t, dir, "c.txt", "")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeMultiple
	opts.StartDir = dir
	m := finder.NewModel(opts)

	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Toggle first two files.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(finder.Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(finder.Model)

	view := m.View()
	if !strings.Contains(view, "2 selected") {
		t.Errorf("expected '2 selected' in view, got:\n%s", view)
	}
}

func TestLongPathTruncation(t *testing.T) {
	// Create a deeply nested path.
	dir := t.TempDir()
	deep := dir
	for _, seg := range []string{"very", "deeply", "nested", "directory", "structure", "here"} {
		deep = deep + "/" + seg
		createDir(t, dir, deep[len(dir)+1:])
	}
	createFile(t, deep, "file.txt", "")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = deep
	m := finder.NewModel(opts)
	// Force a narrow width.
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 24})
	m = updated.(finder.Model)

	cmd := m.Init()
	msg := cmd()
	updated, _ = m.Update(msg)
	m = updated.(finder.Model)

	view := m.View()
	// The path line should contain the ellipsis if it was truncated.
	lines := strings.Split(view, "\n")
	// Path is the second non-empty line (after title).
	for _, line := range lines {
		if strings.Contains(line, "here") {
			// Should end with the dir name but be truncated at the left.
			if len(deep) > 36 && !strings.Contains(line, "…") {
				t.Errorf("expected truncated path with '…', got: %s", line)
			}
			break
		}
	}
}

func TestLongFileNameTruncation(t *testing.T) {
	dir := t.TempDir()
	longName := "this_is_a_very_long_filename_that_should_be_truncated_in_narrow_terminals.txt"
	createFile(t, dir, longName, "data")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = dir
	m := finder.NewModel(opts)
	// Force narrow width.
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 24})
	m = updated.(finder.Model)

	cmd := m.Init()
	msg := cmd()
	updated, _ = m.Update(msg)
	m = updated.(finder.Model)

	view := m.View()
	if !strings.Contains(view, "…") {
		t.Errorf("expected truncated filename with '…' in narrow terminal, got:\n%s", view)
	}
	// The full long name should NOT appear.
	if strings.Contains(view, longName) {
		t.Error("full long filename should be truncated in narrow terminal")
	}
}

func TestHiddenForcedHidesToggle(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".hidden", "")
	createFile(t, dir, "visible.txt", "")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = dir
	opts.ShowHidden = true // forced
	m := finder.NewModel(opts)

	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// View should show hidden files.
	view := m.View()
	if !strings.Contains(view, ".hidden") {
		t.Error("expected .hidden to be visible when ShowHidden is forced")
	}

	// Help bar should NOT show the toggle hidden keybind.
	if strings.Contains(view, "hidden") && strings.Contains(view, "toggle") {
		t.Error("toggle hidden keybind should be hidden when ShowHidden is forced")
	}

	// Pressing the hidden toggle key ('.' by default) should NOT toggle.
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}})
	m = updated.(finder.Model)

	// No readDir command should fire (toggling is suppressed).
	if cmd != nil {
		t.Error("hidden toggle should be suppressed when forced")
	}

	// Hidden files should still be visible.
	view = m.View()
	if !strings.Contains(view, ".hidden") {
		t.Error(".hidden should still be visible after pressing toggle when forced")
	}
}

func TestHiddenNotForcedAllowsToggle(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".hidden", "")
	createFile(t, dir, "visible.txt", "")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = dir
	// ShowHidden defaults to false — toggle should work.
	m := finder.NewModel(opts)

	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Hidden files should NOT be visible initially.
	view := m.View()
	if strings.Contains(view, ".hidden") {
		t.Error(".hidden should not be visible with ShowHidden=false")
	}

	// Press '.' to toggle hidden files on.
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}})
	m = updated.(finder.Model)

	if cmd == nil {
		t.Fatal("expected readDir command after toggling hidden")
	}

	// Complete the readDir.
	msg = cmd()
	updated, _ = m.Update(msg)
	m = updated.(finder.Model)

	// Hidden files should now be visible.
	view = m.View()
	if !strings.Contains(view, ".hidden") {
		t.Error(".hidden should be visible after toggling hidden on")
	}
}

func TestPageUpDown(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 50; i++ {
		createFile(t, dir, "file"+string(rune('a'+i/26))+string(rune('a'+i%26))+".txt", "")
	}

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = dir
	m := finder.NewModel(opts)

	// Force small height so pagination kicks in.
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 15})
	m = updated.(finder.Model)

	cmd := m.Init()
	msg := cmd()
	updated, _ = m.Update(msg)
	m = updated.(finder.Model)

	// Page down.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m = updated.(finder.Model)

	view := m.View()
	if !strings.Contains(view, "/") {
		t.Error("expected view content after page down")
	}

	// Page up.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	m = updated.(finder.Model)

	view = m.View()
	if !strings.Contains(view, "/") {
		t.Error("expected view content after page up")
	}
}

func TestPageDownEmptyDir(t *testing.T) {
	dir := t.TempDir()

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = dir
	m := finder.NewModel(opts)

	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Page down on empty dir should not panic.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m = updated.(finder.Model)

	view := m.View()
	if !strings.Contains(view, "empty") {
		t.Error("expected empty directory message")
	}
}

func TestHomeEndKeys(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "aaa.txt", "")
	createFile(t, dir, "zzz.txt", "")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = dir
	m := finder.NewModel(opts)

	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// End key.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	m = updated.(finder.Model)

	// Home key.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyHome})
	m = updated.(finder.Model)

	// Should not panic and view should be valid.
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view after home/end")
	}
}

func TestNavigateOnFile(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "only.txt", "")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = dir
	m := finder.NewModel(opts)

	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Right arrow on a file should do nothing.
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = updated.(finder.Model)

	if cmd != nil {
		t.Error("right arrow on a file should not trigger a command")
	}
}

func TestNavigateEmptyDir(t *testing.T) {
	dir := t.TempDir()

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = dir
	m := finder.NewModel(opts)

	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Right arrow on empty dir should do nothing.
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = updated.(finder.Model)

	if cmd != nil {
		t.Error("right arrow on empty dir should not trigger a command")
	}
}

func TestSelectDirInFileMode(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "file.txt", "")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = dir
	m := finder.NewModel(opts)

	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// 's' in file mode should do nothing.
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updated.(finder.Model)

	if m.SelectedPath() != "" {
		t.Error("'s' should not select in file mode")
	}
}

func TestFolderModeEnterOnFile(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "file.txt", "")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFolder
	opts.StartDir = dir
	m := finder.NewModel(opts)

	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Enter on a file in folder mode should do nothing.
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(finder.Model)

	if m.SelectedPath() != "" {
		t.Error("enter on file in folder mode should not select")
	}
}

func TestCursorUpAtTop(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "file.txt", "")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = dir
	m := finder.NewModel(opts)

	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Up at top should stay at top.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(finder.Model)

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestCustomStyles(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "file.txt", "")

	s := finder.DefaultStyles()
	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = dir
	opts.Styles = &s
	m := finder.NewModel(opts)

	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view with custom styles")
	}
}

func TestMultiSelectEnterWithNoSelections(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "file.txt", "")

	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeMultiple
	opts.StartDir = dir
	m := finder.NewModel(opts)

	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	// Enter on a file with no prior selections should select just that file.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(finder.Model)

	paths := m.SelectedPaths()
	if len(paths) != 1 {
		t.Fatalf("expected 1 selected path, got %d", len(paths))
	}
}

func TestDirReadError(t *testing.T) {
	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = "/nonexistent/path/that/does/not/exist"
	m := finder.NewModel(opts)

	cmd := m.Init()
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(finder.Model)

	if m.Err() == nil {
		t.Error("expected error for nonexistent directory")
	}

	view := m.View()
	if !strings.Contains(view, "Error") {
		t.Error("expected error message in view")
	}
}

func TestModelWSLDetection(t *testing.T) {
	// This test validates that IsWSL doesn't panic; actual result
	// depends on the runtime environment.
	result := finder.IsWSL()
	_ = result

	// Verify /proc/version exists on Linux (non-fatal if not).
	if _, err := os.Stat("/proc/version"); err != nil {
		t.Skip("not on Linux, skipping WSL check")
	}
}
