# go-finder

[![CI](https://github.com/SREsAreHumanToo/go-finder/actions/workflows/ci.yml/badge.svg)](https://github.com/SREsAreHumanToo/go-finder/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/SREsAreHumanToo/go-finder.svg)](https://pkg.go.dev/github.com/SREsAreHumanToo/go-finder)
[![Go Report Card](https://goreportcard.com/badge/github.com/SREsAreHumanToo/go-finder)](https://goreportcard.com/report/github.com/SREsAreHumanToo/go-finder)
[![codecov](https://codecov.io/gh/SREsAreHumanToo/go-finder/graph/badge.svg)](https://codecov.io/gh/SREsAreHumanToo/go-finder)

A cross-platform, terminal-based file and folder picker for Go. Works consistently across Windows, macOS, Linux, BSD, WSL, and Git Bash with zero OS-specific dependencies.

## Features

- **Four picker modes**: single file, single folder, any (file or folder), and multi-select
- **Fuzzy search filtering**: press `/` to filter entries in real time, ranked best-match-first (opt out with `WithFuzzySearch(false)` for plain substring matching)
- **Interactive mode**: create files/folders and delete entries without leaving the picker
- **Glob-based file filtering** (`*.go`, `*.txt`, etc.)
- **Hidden file support**: toggle at runtime, or force-show with distinct dim styling
- **Symlink expansion**: optionally resolve symlinks to real paths
- **Pluggable filesystem**: browse any `io/fs.FS` (e.g. `embed.FS`, `fstest.MapFS`) via `WithFS` — defaults to the host OS
- **Preview pane**: optional side pane (`WithPreview`) showing a file's head, a directory's listing, or metadata; customizable via `WithPreviewFunc`
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
| Type characters | Filter entries live (fuzzy, ranked best-match-first) |
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

# Interactive mode (create/delete)
go run ./examples/interactive

# Custom keybindings and styles
go run ./examples/custom
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
