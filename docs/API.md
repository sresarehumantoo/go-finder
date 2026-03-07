# API Reference

## Public Functions

### `PickFile(opts ...Option) (string, error)`

Opens an interactive terminal-based file picker. Returns the absolute path of the selected file, or `ErrCancelled` if the user quits without selecting.

```go
path, err := finder.PickFile(
    finder.WithStartDir("/home/user/projects"),
    finder.WithFilter("*.go", "*.mod"),
)
```

### `PickFolder(opts ...Option) (string, error)`

Opens a folder picker. Only directories can be selected. Press `enter` on a highlighted directory to select it, or `s` to select the current working directory. Use `right`/`l` to navigate into directories.

```go
dir, err := finder.PickFolder()
```

### `PickAny(opts ...Option) (string, error)`

Opens a picker that allows selecting either a file or folder. Press `enter` to open directories or select files, `tab` to select the highlighted item directly, or `s` to select the current directory.

```go
path, err := finder.PickAny()
```

### `PickMultiple(opts ...Option) ([]string, error)`

Opens a multi-select picker. Use `space`/`tab` to toggle selection on individual items, `ctrl+a` to toggle all, `enter` on a directory to navigate into it (selections persist across directories), and `enter` on a file to confirm. The status bar shows the current selection count.

```go
paths, err := finder.PickMultiple(
    finder.WithFilter("*.log"),
    finder.WithHidden(true),
)
```

## Options

All option functions follow the functional options pattern.

| Function | Description | Default |
|---|---|---|
| `WithStartDir(dir string)` | Set starting directory | Current working directory |
| `WithFilter(patterns ...string)` | Glob patterns to filter files (dirs always shown) | Show all files |
| `WithHidden(show bool)` | Show hidden files (force-on disables toggle key) | `false` |
| `WithTitle(title string)` | Header text | Mode-dependent |
| `WithHeight(h int)` | Max visible rows (0 = auto) | `0` |
| `WithMode(m Mode)` | Picker mode | Set by Pick function |
| `WithInteractive(enabled bool)` | Enable create/delete actions | `false` |
| `WithExpandSymlinks(enabled bool)` | Resolve symlinks to real paths | `false` |
| `WithKeyMap(km KeyMap)` | Override default keybindings | Default key map |
| `WithStyles(s Styles)` | Override default visual styles | Default styles |

## Types

### `Mode`

```go
const (
    ModeFile     Mode = iota  // Single file selection
    ModeFolder                // Single folder selection
    ModeAny                   // Select either a file or folder
    ModeMultiple              // Multi-select (files and/or folders across directories)
)
```

### `FileEntry`

```go
type FileEntry struct {
    Name     string
    Path     string
    IsDir    bool
    IsHidden bool
    Size     int64
    Mode     os.FileMode
}
```

### `KeyMap`

All keybindings are fully overridable. Each field is a `key.Binding` from `charmbracelet/bubbles/key`.

```go
type KeyMap struct {
    Up, Down, PageUp, PageDown  key.Binding
    Home, End                   key.Binding
    Navigate, Back              key.Binding
    Select, SelectDir           key.Binding
    Toggle, ToggleAll           key.Binding
    Hidden, Cancel, Search      key.Binding
    NewFile, NewFolder, Delete  key.Binding
}
```

Example — rebind cancel to `x`:

```go
km := finder.DefaultKeyMap()
km.Cancel = key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "exit"))
path, err := finder.PickFile(finder.WithKeyMap(km))
```

### `Styles`

All visual styles are overridable via `lipgloss.Style` values.

```go
type Styles struct {
    Title, Path, Cursor, Selected          lipgloss.Style
    Directory, HiddenDir                    lipgloss.Style
    File, HiddenFile, FileSize, Permission lipgloss.Style
    StatusBar, SearchPrompt, SearchText    lipgloss.Style
    Help, HelpKey, HelpDesc, HelpSep       lipgloss.Style
    Border                                  lipgloss.Style
}
```

### Errors

- `ErrCancelled` — Returned when the user exits the picker without making a selection.

## Utility Functions

### Path Conversion (WSL)

```go
// Convert WSL path to Windows path
winPath := finder.ToWindowsPath("/mnt/c/Users/test")  // "C:\Users\test"

// Convert Windows path to WSL path
wslPath := finder.ToWSLPath(`C:\Users\test`)  // "/mnt/c/Users/test"

// Detect if running in WSL
if finder.IsWSL() {
    // handle WSL-specific logic
}
```

### Filesystem

```go
// Read directory contents with filtering
entries, err := finder.ReadDir("/path/to/dir", showHidden, []string{"*.go"})

// Get parent directory
parent := finder.ParentDir("/home/user/projects")  // "/home/user"

// Normalize and resolve a path
abs := finder.NormalizePath("~/projects/../code")

// Resolve symlinks
real := finder.ResolvePath("/path/to/symlink")

// Human-readable file size
size := finder.FormatSize(1536)  // "1.5 KB"

// Create file or directory
err := finder.CreateFile("/path/to/dir", "newfile.txt")
err := finder.CreateDir("/path/to/dir", "newdir")

// Delete file or directory (recursive)
err := finder.DeletePath("/path/to/target")
```

## Behavior Notes

### Hidden files with `-hidden`

When `ShowHidden` is set to `true` at construction (e.g. via `WithHidden(true)`), the toggle hidden keybind (`.`) is disabled and removed from the help bar. Hidden files are always shown with a distinct dim italic style to differentiate them from normal entries.

### Search filtering

Pressing `/` enters search mode. As you type, entries that don't match are hidden in real time. Press `enter` to accept the filtered view, or `esc` to cancel and restore the full list. Search is case-insensitive.

### Trailing slash creates directory

When using interactive mode (`n` to create a new file), typing a name ending with `/` (e.g. `config/`) creates a directory instead of a file.

### Multi-select across directories

In multi-select mode, selections are preserved when navigating between directories. The status bar shows the total selection count. Press `enter` on a directory to navigate into it, `enter` on a file to confirm all selections.

### Long path/name truncation

Paths in the header are truncated from the left with `…` when they exceed terminal width. File and directory names in the list are truncated from the right with `…` to prevent line overflow.
