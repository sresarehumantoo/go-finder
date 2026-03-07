// Package finder provides a cross-platform, terminal-based file and folder
// picker that works consistently across Windows, macOS, Linux, BSD, WSL, and
// Git Bash without any OS-specific dependencies.
package finder

import (
	"os"
	"path/filepath"
)

// Mode defines the picker operation mode.
type Mode int

const (
	// ModeFile allows selecting a single file.
	ModeFile Mode = iota
	// ModeFolder allows selecting a single folder.
	ModeFolder
	// ModeAny allows selecting a single file or folder.
	ModeAny
	// ModeMultiple allows selecting multiple files and/or folders.
	ModeMultiple
)

// String returns the human-readable name of the mode.
func (m Mode) String() string {
	switch m {
	case ModeFile:
		return "file"
	case ModeFolder:
		return "folder"
	case ModeAny:
		return "any"
	case ModeMultiple:
		return "multiple"
	default:
		return "unknown"
	}
}

// Options holds the configuration for a picker session.
type Options struct {
	// StartDir is the initial directory shown in the picker.
	StartDir string
	// Filters is a list of glob patterns to filter visible files (e.g. "*.go", "*.txt").
	// Empty means show all files.
	Filters []string
	// ShowHidden controls whether dotfiles/dotfolders are displayed.
	ShowHidden bool
	// Mode determines the picker behavior (file, folder, or multi-select).
	Mode Mode
	// Title is the header text displayed at the top of the picker.
	Title string
	// Height constrains the maximum number of rows in the file list.
	// Zero means auto-detect from terminal size.
	Height int
	// Interactive enables file management actions (create, delete).
	Interactive bool
	// ExpandSymlinks resolves symlinks to their real paths. When enabled,
	// navigating back from a symlinked directory goes to the real parent
	// rather than the symlink's parent.
	ExpandSymlinks bool
	// KeyMap overrides the default keybindings. Nil means use defaults.
	KeyMap *KeyMap
	// Styles overrides the default visual styles. Nil means use defaults.
	Styles *Styles
}

// Option is a functional option for configuring the picker.
type Option func(*Options)

// DefaultOptions returns a new Options struct with sensible defaults.
func DefaultOptions() Options {
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}
	dir, _ = filepath.Abs(dir)

	return Options{
		StartDir:   dir,
		Filters:    nil,
		ShowHidden: false,
		Mode:       ModeFile,
		Title:      "Select a file",
		Height:     0,
	}
}

// WithStartDir sets the initial directory for the picker.
func WithStartDir(dir string) Option {
	return func(o *Options) {
		abs, err := filepath.Abs(dir)
		if err == nil {
			o.StartDir = abs
		} else {
			o.StartDir = dir
		}
	}
}

// WithFilter adds glob patterns to limit which files are shown.
// Directories are always shown regardless of filters.
func WithFilter(patterns ...string) Option {
	return func(o *Options) {
		o.Filters = append(o.Filters, patterns...)
	}
}

// WithHidden controls whether hidden files and directories are shown.
func WithHidden(show bool) Option {
	return func(o *Options) {
		o.ShowHidden = show
	}
}

// WithTitle sets the header text displayed at the top of the picker.
func WithTitle(title string) Option {
	return func(o *Options) {
		o.Title = title
	}
}

// WithHeight sets the maximum visible rows for the file list.
// A value of 0 means auto-detect from terminal size.
func WithHeight(h int) Option {
	return func(o *Options) {
		o.Height = h
	}
}

// WithMode sets the picker mode (file, folder, or multi-select).
func WithMode(m Mode) Option {
	return func(o *Options) {
		o.Mode = m
	}
}

// WithInteractive enables file management actions (create files/folders, delete).
func WithInteractive(enabled bool) Option {
	return func(o *Options) {
		o.Interactive = enabled
	}
}

// WithExpandSymlinks resolves symlinks to their real paths. When enabled,
// the picker operates on resolved paths so navigating back from a symlinked
// directory goes to the real parent rather than the symlink's parent.
func WithExpandSymlinks(enabled bool) Option {
	return func(o *Options) {
		o.ExpandSymlinks = enabled
	}
}

// WithKeyMap overrides the default keybindings. All fields of the KeyMap
// struct are public, so callers can customize any or all bindings.
func WithKeyMap(km KeyMap) Option {
	return func(o *Options) {
		o.KeyMap = &km
	}
}

// WithStyles overrides the default visual styles.
func WithStyles(s Styles) Option {
	return func(o *Options) {
		o.Styles = &s
	}
}
