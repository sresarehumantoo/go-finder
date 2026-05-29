# go-finder

> A **TUI file picker, file explorer, and fuzzy finder for Go!** Browse, search, and select files or folders from the terminal like the ultimate superuser you are.

[![CI](https://github.com/SREsAreHumanToo/go-finder/actions/workflows/ci.yml/badge.svg)](https://github.com/SREsAreHumanToo/go-finder/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/SREsAreHumanToo/go-finder.svg)](https://pkg.go.dev/github.com/SREsAreHumanToo/go-finder)
[![Go Report Card](https://goreportcard.com/badge/github.com/SREsAreHumanToo/go-finder)](https://goreportcard.com/report/github.com/SREsAreHumanToo/go-finder)
[![codecov](https://codecov.io/gh/SREsAreHumanToo/go-finder/graph/badge.svg)](https://codecov.io/gh/SREsAreHumanToo/go-finder)

A cross-platform, terminal-based (TUI) file and folder picker, explorer, and fuzzy finder for Go. Works consistently across Windows, macOS, Linux, BSD, WSL, and Git Bash with zero OS-specific dependencies.

```go
path, err := finder.PickFile()   // one import, one line
```

Most Go terminal pickers are *components* you wire into your own Bubble Tea
update/view loop. **go-finder is the batteries-included alternative**: a single
call that returns the selected path — with fuzzy search, multi-select, a preview
pane, and interactive create/delete built in — yet it still embeds as a Bubble
Tea sub-model when you want full control.

## Why go-finder?

| | go-finder | [bubbles/filepicker](https://github.com/charmbracelet/bubbles) | [huh](https://github.com/charmbracelet/huh) file field | [go-fuzzyfinder](https://github.com/ktr0731/go-fuzzyfinder) |
|---|:---:|:---:|:---:|:---:|
| One-line standalone API | ✅ | ❌ | ❌ | ✅ |
| Embeddable Bubble Tea sub-model | ✅ | ✅ | ➖ form field | ❌ |
| Directory tree navigation | ✅ | ✅ | ✅ | ❌ list only |
| Fuzzy search + highlight | ✅ | ❌ | ❌ | ✅ |
| Multi-select | ✅ | ❌ | ❌ | ✅ |
| File-or-folder (`ModeAny`) | ✅ | ➖ | ❌ | ❌ |
| Interactive create / delete | ✅ | ❌ | ❌ | ❌ |
| Preview pane | ✅ | ❌ | ❌ | ✅ |
| Custom `io/fs.FS` backend | ✅ | ❌ | ❌ | ➖ |

See [`docs/POSITIONING.md`](docs/POSITIONING.md) for the full landscape analysis.

## Features

- **Four picker modes**: single file, single folder, any (file or folder), and multi-select
- **Fuzzy search filtering**: press `/` to filter entries in real time, ranked best-match-first with matched characters highlighted (opt out with `WithFuzzySearch(false)` for plain substring matching)
- **Interactive mode**: create files/folders and delete entries without leaving the picker
- **Glob-based file filtering** (`*.go`, `*.txt`, etc.)
- **Hidden file support**: toggle at runtime, or force-show with distinct dim styling
- **Symlink expansion**: optionally resolve symlinks to real paths
- **Pluggable filesystem**: browse any `io/fs.FS` (e.g. `embed.FS`, `fstest.MapFS`) via `WithFS` — defaults to the host OS
- **Preview pane**: optional side pane (`WithPreview`) showing a file's head, a directory's listing, or metadata; customizable via `WithPreviewFunc`
- **Standalone or embeddable**: one-line `PickFile()` etc., or embed the picker as a Bubble Tea sub-model with `New` (reports completion via `DoneMsg`)
- **Smart truncation**: long paths and filenames are truncated with `…` to fit the terminal
- **Fully customizable**: override any keybinding or visual style
- **Vim-style navigation** (`h/j/k/l`) plus standard arrow keys
- **WSL path conversion** utilities (`/mnt/c/...` ↔ `C:\...`)
- **Alt-screen mode**: restores terminal on exit
- **No cgo**, no external dependencies beyond pure-Go TUI libraries

## Install

```bash
go get github.com/SREsAreHumanToo/go-finder@latest
```

> The module path is case-sensitive: use `SREsAreHumanToo` exactly as shown.
> Go's module proxy distinguishes capitalization, so `sresarehumantoo`
> will fail with a "case mismatch" error.

## Quick Start

```go
package main

import (
    "fmt"
    finder "github.com/SREsAreHumanToo/go-finder"
)

func main() {
    path, err := finder.PickFile(
        finder.WithStartDir("~/projects"),
        finder.WithFilter("*.go"),
    )
    if err != nil {
        panic(err)
    }
    fmt.Println("Selected:", path)
}
```

### Pick a folder

```go
dir, err := finder.PickFolder()
```

### Pick a file or folder

```go
path, err := finder.PickAny()
```

### Multi-select

```go
paths, err := finder.PickMultiple(
    finder.WithFilter("*.log", "*.txt"),
    finder.WithHidden(true),
)
```

### Restrict to document types

`WithExtensions` limits the picker to a set of extensions, matched
case-insensitively (so `Report.PDF` is included). Use it instead of
`WithFilter("*.pdf")` when you want the common "only allow these file types"
behavior without worrying about case:

```go
path, err := finder.PickFile(
    finder.WithExtensions("pdf", "docx", "doc"),
)
```

### Interactive mode (create/delete)

```go
path, err := finder.PickFile(
    finder.WithInteractive(true),
)
```

### Custom keybindings and styles

```go
km := finder.DefaultKeyMap()
km.Cancel = key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "exit"))

s := finder.DefaultStyles()
s.Directory = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)

path, err := finder.PickFile(
    finder.WithKeyMap(km),
    finder.WithStyles(s),
)
```

### Embed in your own Bubble Tea app

The `Pick*` functions run a standalone program. To embed the picker as a
sub-model in an existing Bubble Tea program, use `New`, which returns a model in
embedded mode: it never quits your program and reports completion via
`finder.DoneMsg`.

```go
type parent struct{ picker finder.Model }

func (m parent) Init() tea.Cmd { return m.picker.Init() }

func (m parent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if done, ok := msg.(finder.DoneMsg); ok {
        if done.State == finder.StateSelected {
            // use done.Paths
        }
        return m, tea.Quit // or switch to another view
    }
    updated, cmd := m.picker.Update(msg)
    m.picker = updated.(finder.Model)
    return m, cmd
}

func (m parent) View() string { return m.picker.View() }

// finder.New(finder.WithMode(finder.ModeFile), finder.WithPreview(true))
```

See [`examples/embedded`](examples/embedded) for a complete program.

## Keybindings

### Navigation

| Key | Action |
|---|---|
| `up` / `k` | Move cursor up |
| `down` / `j` | Move cursor down |
| `pgup` / `ctrl+u` | Page up |
| `pgdn` / `ctrl+d` | Page down |
| `g` / `home` | Jump to top |
| `G` / `end` | Jump to bottom |
| `right` / `l` | Open directory |
| `backspace` / `left` / `h` / `esc` | Go to parent directory |

### Selection

| Key | Action | Modes |
|---|---|---|
| `enter` | Select file / open directory | All modes |
| `space` / `tab` | Toggle selection / select item | Multi-select, Any |
| `ctrl+a` | Toggle all selections | Multi-select |
| `s` | Select current directory | Folder, Any |

### Other

| Key | Action |
|---|---|
| `/` | Search / filter entries (live) |
| `.` | Toggle hidden files |
| `n` | New file (interactive mode) |
| `N` | New folder (interactive mode) |
| `d` / `delete` | Delete entry (interactive mode) |
| `q` / `ctrl+c` | Cancel and exit |

### Search mode

| Key | Action |
|---|---|
| Type characters | Filter entries live (fuzzy, ranked, matches highlighted) |
| `backspace` | Remove last character (widens results) |
| `enter` | Accept filtered results |
| `esc` | Cancel search, restore full list |

### Interactive input (new file/folder)

| Key | Action |
|---|---|
| Type characters | Enter name |
| `enter` | Create the file or folder |
| `esc` | Cancel |

Typing a name ending with `/` when creating a file (e.g. `mydir/`) creates a directory instead.

## Examples

```bash
# File picker (all flags)
go run ./examples/basic

# Folder picker
go run ./examples/folder

# Multi-select with filters
go run ./examples/multi

# Restrict to file extensions (case-insensitive, opt-in)
go run ./examples/extensions

# Interactive mode (create/delete)
go run ./examples/interactive

# Custom keybindings and styles
go run ./examples/custom

# Preview pane (built-in or custom)
go run ./examples/preview

# Embedded as a Bubble Tea sub-model
go run ./examples/embedded
```

The basic example supports flags for all options:

```bash
go run ./examples/basic -mode folder
go run ./examples/basic -mode multi -filter "*.go"
go run ./examples/basic -hidden
go run ./examples/basic -interactive
go run ./examples/basic -dir /tmp
go run ./examples/basic -expand -dir ~/symlink
```

## Documentation

- [API Reference](docs/API.md)
- [Architecture](docs/ARCHITECTURE.md)

## Requirements

- Go 1.25+
- A terminal with ANSI escape code support (virtually all modern terminals)

## License

MIT - see [LICENSE](LICENSE) for details.
