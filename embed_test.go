package finder_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	finder "github.com/rummage-dev/rummage"
)

// loadEmbedded builds an embedded picker over dir and loads its entries.
func loadEmbedded(t *testing.T, dir string, opts ...finder.Option) finder.Model {
	t.Helper()
	all := append([]finder.Option{finder.WithStartDir(dir)}, opts...)
	m := finder.New(all...)
	updated, _ := m.Update(m.Init()())
	return updated.(finder.Model)
}

// drainDone runs cmd and returns the DoneMsg it produced, if any.
func drainDone(t *testing.T, cmd tea.Cmd) (finder.DoneMsg, bool) {
	t.Helper()
	if cmd == nil {
		return finder.DoneMsg{}, false
	}
	msg := cmd()
	done, ok := msg.(finder.DoneMsg)
	return done, ok
}

func TestEmbeddedSelectEmitsDoneMsgNotQuit(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "target.txt", "data")

	m := loadEmbedded(t, dir, finder.WithMode(finder.ModeFile))
	if m.State() != finder.StateBrowsing {
		t.Fatalf("expected StateBrowsing before selection, got %v", m.State())
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(finder.Model)

	if m.State() != finder.StateSelected {
		t.Errorf("expected StateSelected after enter, got %v", m.State())
	}
	if !m.Done() {
		t.Error("expected Done() to be true after selection")
	}
	if got := m.SelectedPath(); got == "" {
		t.Error("expected a selected path")
	}

	done, ok := drainDone(t, cmd)
	if !ok {
		t.Fatal("expected a DoneMsg from the returned command")
	}
	if done.State != finder.StateSelected {
		t.Errorf("expected DoneMsg.State StateSelected, got %v", done.State)
	}
	if len(done.Paths) != 1 {
		t.Errorf("expected one path in DoneMsg, got %v", done.Paths)
	}
	// The command must not be tea.Quit (which would kill the parent program).
	if _, isQuit := cmd().(tea.QuitMsg); isQuit {
		t.Error("embedded picker must not return tea.Quit on selection")
	}
}

func TestEmbeddedCancel(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "a.txt", "x")

	m := loadEmbedded(t, dir)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updated.(finder.Model)

	if m.State() != finder.StateCancelled {
		t.Errorf("expected StateCancelled after q, got %v", m.State())
	}
	if got := m.SelectedPath(); got != "" {
		t.Errorf("expected no selection after cancel, got %q", got)
	}
	done, ok := drainDone(t, cmd)
	if !ok || done.State != finder.StateCancelled {
		t.Errorf("expected DoneMsg{StateCancelled}, got %+v ok=%v", done, ok)
	}
}

func TestEmbeddedMultiSelect(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "a.txt", "x")
	createFile(t, dir, "b.txt", "x")

	m := loadEmbedded(t, dir, finder.WithMode(finder.ModeMultiple))

	// Toggle both files, then confirm with enter.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(finder.Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(finder.Model)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(finder.Model)

	if m.State() != finder.StateSelected {
		t.Fatalf("expected StateSelected, got %v", m.State())
	}
	done, ok := drainDone(t, cmd)
	if !ok {
		t.Fatal("expected DoneMsg")
	}
	if len(done.Paths) != 2 {
		t.Errorf("expected 2 selected paths, got %v", done.Paths)
	}
}

func TestEmbeddedViewRendersAfterDone(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "target.txt", "data")

	m := loadEmbedded(t, dir, finder.WithMode(finder.ModeFile))
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(finder.Model)

	// Embedded picker does not blank its view on completion; the parent
	// controls when to stop rendering it.
	if m.View() == "" {
		t.Error("embedded picker should still render its view after selection")
	}
}

func TestStandaloneStillSelects(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "target.txt", "data")

	// NewModel without Embedded is the standalone path used by Pick*.
	opts := finder.DefaultOptions()
	opts.Mode = finder.ModeFile
	opts.StartDir = dir
	m := finder.NewModel(opts)
	updated, _ := m.Update(m.Init()())
	m = updated.(finder.Model)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(finder.Model)

	if m.State() != finder.StateSelected {
		t.Errorf("expected StateSelected, got %v", m.State())
	}
	if m.SelectedPath() == "" {
		t.Error("expected a selected path in standalone mode")
	}
	// Standalone selection ends the program via tea.Quit.
	if cmd == nil {
		t.Fatal("expected a command from standalone selection")
	}
	if _, isQuit := cmd().(tea.QuitMsg); !isQuit {
		t.Error("standalone picker should return tea.Quit on selection")
	}
}
