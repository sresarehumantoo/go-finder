// Package finder provides a cross-platform, terminal-based file and folder
// picker that works consistently across Windows, macOS, Linux, BSD, WSL, and
// Git Bash without any OS-specific dependencies.
//
// The picker is built on the Bubble Tea TUI framework and uses no cgo, so it
// cross-compiles cleanly and runs in any environment where a terminal is
// available — including SSH sessions and Git Bash, where native OS file
// dialogs are not usable.
//
// # Quick Start
//
// The four Pick functions are the most common entry points:
//
//	path, err := finder.PickFile(finder.WithStartDir("~/projects"))
//	dir, err  := finder.PickFolder()
//	any, err  := finder.PickAny()
//	paths, err := finder.PickMultiple(finder.WithFilter("*.go"))
//
// Each returns [ErrCancelled] if the user exits the picker without making a
// selection. Configuration is applied via functional [Option] values.
//
// # Modes
//
// The [Mode] constants control selection behavior:
//
//   - [ModeFile]     — select a single file
//   - [ModeFolder]   — select a single folder
//   - [ModeAny]      — select either a file or folder
//   - [ModeMultiple] — select multiple files and/or folders across directories
//
// # Customization
//
// Every keybinding and visual style is overridable through [WithKeyMap] and
// [WithStyles]. See [DefaultKeyMap] and [DefaultStyles] for the defaults.
//
// # Interactive Mode
//
// Passing [WithInteractive] enables in-picker file management: press n to
// create a file, N to create a folder, and d to delete an entry. Typing a
// new file name ending in "/" creates a directory instead.
//
// # Utilities
//
// The package also exposes path utilities that are useful outside the picker
// itself, including [ToWindowsPath], [ToWSLPath], [IsWSL], [ResolvePath],
// [NormalizePath], [ParentDir], and [FormatSize].
package finder
