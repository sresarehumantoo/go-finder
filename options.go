package finder

import (
	"io/fs"
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
	// FuzzySearch enables scored fuzzy matching for the search filter,
	// ranking results best-match-first. When false, search uses a plain
	// case-insensitive substring match that preserves the original order.
	FuzzySearch bool
	// Preview enables a right-hand pane that previews the highlighted entry
	// (file head, directory listing, or metadata). Off by default.
	Preview bool
	// PreviewFunc customizes how the highlighted entry is previewed. When nil,
	// a built-in preview is used. It receives the entry and the pane's
	// width/height (in cells) and returns the preview text; the result is
	// clipped to fit the pane.
	PreviewFunc PreviewFunc
	// FS is the filesystem backend the picker reads from. Nil means the
	// host operating system. Set it via WithFS to browse a custom io/fs.FS
	// (e.g. an embed.FS or fstest.MapFS); such filesystems are read-only,
	// so interactive create/delete is disabled for them.
	FS FileSystem
	// KeyMap overrides the default keybindings. Nil means use defaults.
	KeyMap *KeyMap
	// Styles overrides the default visual styles. Nil means use defaults.
	Styles *Styles
}

// Option is a functional option for configuring the picker.
type Option func(*Options)

// PreviewFunc renders a preview of the highlighted entry. It is given the
// entry and the preview pane's width and height in terminal cells, and returns
// the preview text (which may span multiple lines). The returned text is
// clipped to the pane's dimensions.
type PreviewFunc func(e FileEntry, width, height int) string

// DefaultOptions returns a new Options struct with sensible defaults.
func DefaultOptions() Options {
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}
	dir, _ = filepath.Abs(dir)

	return Options{
		StartDir:    dir,
		Filters:     nil,
		ShowHidden:  false,
		Mode:        ModeFile,
		Title:       "Select a file",
		Height:      0,
		FuzzySearch: true,
	}
}

// WithStartDir sets the initial directory for the picker. For the default OS
// backend the path is resolved to an absolute path when the picker starts; for
// a custom filesystem (see WithFS) it is interpreted as a relative,
// slash-separated path within that filesystem.
func WithStartDir(dir string) Option {
	return func(o *Options) {
		o.StartDir = dir
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

// WithPreview enables (or disables) the preview pane shown beside the file
// list. The pane is hidden automatically when the terminal is too narrow.
func WithPreview(enabled bool) Option {
	return func(o *Options) {
		o.Preview = enabled
	}
}

// WithPreviewFunc sets a custom preview renderer and enables the preview pane.
// Pass nil to fall back to the built-in preview while keeping the pane enabled.
func WithPreviewFunc(fn PreviewFunc) Option {
	return func(o *Options) {
		o.PreviewFunc = fn
		o.Preview = true
	}
}

// WithFS sets a custom filesystem backend to browse instead of the host OS.
// The picker navigates the given io/fs.FS in slash-separated path space,
// starting at its root ("."). Because io/fs is read-only, interactive
// create/delete is unavailable on a custom filesystem.
//
// To start in a subdirectory, combine with WithStartDir using a relative,
// slash-separated path (e.g. WithStartDir("docs/api")).
func WithFS(fsys fs.FS) Option {
	return func(o *Options) {
		o.FS = fsAdapter{fsys: fsys, label: "/"}
	}
}

// WithFuzzySearch toggles scored fuzzy matching for the search filter.
// When enabled (the default), results are ranked best-match-first. When
// disabled, search falls back to a case-insensitive substring match that
// preserves the original directory ordering.
func WithFuzzySearch(enabled bool) Option {
	return func(o *Options) {
		o.FuzzySearch = enabled
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
