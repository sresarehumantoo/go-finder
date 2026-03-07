# Architecture

## Overview

go-finder is a cross-platform, terminal-based file and folder picker built on the [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework. It provides a simple blocking API that spawns an interactive terminal UI and returns the user's selection.

## Design Decisions

### Why TUI instead of native dialogs?

Native file dialogs (Windows `IFileDialog`, macOS `NSOpenPanel`, Linux `zenity`/`kdialog`) require either cgo bindings or shelling out to OS-specific tools. This creates several problems:

1. **Cross-compilation breaks** with cgo
2. **WSL is unreliable** when calling PowerShell for dialogs
3. **Git Bash / SSH sessions** have no display server
4. **Dependency on external tools** (zenity may not be installed)

A TUI picker works identically everywhere a terminal exists.

### Elm Architecture (Bubble Tea)

The picker follows the Elm architecture via Bubble Tea:

1. **Model** (`model.go`) - Holds all state: current directory, file list, cursor position, selections, search term, input mode
2. **Update** (`model.go`) - Receives a `tea.Msg` (keyboard input, directory-read result), processes it, returns new state and optional command
3. **View** (`view.go`) - Pure function that takes model state and renders it to a string for display

The cycle is: user input produces a `tea.Msg`, `Update` returns new state, `View` renders the new state.

### File Layout

```
finder.go         Public API (PickFile, PickFolder, PickAny, PickMultiple)
options.go        Configuration types and functional options
model.go          Bubble Tea model, Init, Update logic
view.go           Rendering / View function
keys.go           Keybinding definitions (fully overridable)
styles.go         Lipgloss style definitions (fully overridable)
filesystem.go     OS-agnostic file operations, path utilities

*_test.go         Tests (one per concern area, package finder_test)

examples/
  basic/          Full-featured demo with all flags
  folder/         Simple folder picker
  multi/          Multi-select with filters
  interactive/    Create/delete file management
  custom/         Custom keybindings and styles

docs/
  API.md          Full API reference
  ARCHITECTURE.md This file
```

### Separation of Concerns

| Layer | Responsibility |
|---|---|
| `finder.go` | Public API surface. Constructs model, runs Bubble Tea program, extracts results |
| `options.go` | Configuration. No logic, just data and option functions |
| `model.go` | All state transitions. Keyboard handling, cursor movement, directory navigation, search filtering, interactive input |
| `view.go` | Pure rendering. Takes model state, returns string. Handles truncation and layout |
| `keys.go` | Keybinding declarations only |
| `styles.go` | Visual styling declarations only (includes hidden file styles) |
| `filesystem.go` | All OS interaction. Directory reading, path conversion, WSL detection, file/dir creation and deletion |

### Key State Management

| State | Purpose |
|---|---|
| `entries` / `allEntries` | Current (possibly filtered) file list / unfiltered backup during search |
| `cursor` / `offset` | Current selection position and scroll viewport |
| `selected` | Map of toggled paths (multi-select, persists across directories) |
| `searching` / `searchTerm` | Live search mode and current filter text |
| `inputMode` / `inputText` | Interactive create/delete prompt state |
| `returnTo` | Remembers child dir name when navigating back (cursor memory) |
| `hiddenForced` | Tracks whether ShowHidden was set at construction (disables toggle) |

### Async Directory Reading

Directory reads are performed as Bubble Tea commands (async). This prevents the UI from blocking on slow filesystems (network drives, large directories).

```go
func (m Model) readDir() tea.Cmd {
    return func() tea.Msg {
        entries, err := ReadDir(m.dir, m.options.ShowHidden, m.options.Filters)
        return dirReadMsg{entries: entries, dir: m.dir, err: err}
    }
}
```

### Mode Behavior Matrix

| Action | ModeFile | ModeFolder | ModeAny | ModeMultiple |
|---|---|---|---|---|
| `enter` on dir | Navigate into | Select dir | Navigate into | Navigate into |
| `enter` on file | Select file | - | Select file | Confirm selections |
| `right`/`l` on dir | Navigate into | Navigate into | Navigate into | Navigate into |
| `tab`/`space` | - | - | Select item | Toggle selection |
| `ctrl+a` | - | - | - | Toggle all |
| `s` | - | Select current dir | Select current dir | - |

### Display Features

- **Path truncation**: Long directory paths in the header are truncated from the left with `…`
- **Name truncation**: Long file/directory names are truncated from the right with `…` based on available width
- **Hidden file styling**: Hidden entries (dotfiles) use dim italic styling when visible
- **Selection count**: Multi-select mode shows "N selected" in the status bar
- **Scroll indicator**: Shows position (e.g. "5/23") when entries exceed visible rows
- **Dynamic help bar**: Wraps keybinding hints based on terminal width, hides irrelevant bindings

## Platform Notes

| Platform | Status | Notes |
|---|---|---|
| Linux | Full support | Primary development target |
| macOS | Full support | Terminal.app and iTerm2 tested |
| Windows | Full support | Windows Terminal recommended |
| WSL | Full support | Path conversion utilities included |
| Git Bash | Full support | ANSI escape support required (mintty OK) |
| BSD | Full support | Same as Linux path |
| cmd.exe | Partial | Older versions may lack ANSI support |
