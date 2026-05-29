package finder

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// FileSystem is the backend the picker reads from and navigates. It abstracts
// both directory access and path manipulation so the picker can operate on the
// OS filesystem (the default) or on any custom backend such as an embed.FS or
// a fstest.MapFS supplied via WithFS.
//
// All path-shaped values the picker stores (the current directory and each
// FileEntry.Path) are produced by this interface, so the picker never assumes
// OS path semantics.
type FileSystem interface {
	// Open opens the named file for reading. This makes a FileSystem usable
	// anywhere an fs.FS is expected (e.g. the preview pane).
	Open(name string) (fs.File, error)
	// ReadDir returns the raw directory entries at the given path.
	ReadDir(name string) ([]fs.DirEntry, error)
	// Parent returns the parent directory of path. For the root it returns
	// path unchanged, which is how the picker detects the top level.
	Parent(path string) string
	// Base returns the final element of path.
	Base(path string) string
	// Join appends name to dir, producing a child path in this filesystem's
	// path space.
	Join(dir, name string) string
	// Display renders path for the breadcrumb shown to the user.
	Display(path string) string
	// Root returns the path the picker starts at when no start directory is
	// given.
	Root() string
}

// WritableFileSystem is the optional extension a FileSystem implements to
// support interactive mode (creating and deleting entries). A FileSystem that
// does not implement it is treated as read-only: the create/delete keys show a
// status message instead of acting.
type WritableFileSystem interface {
	FileSystem
	CreateFile(dir, name string) error
	CreateDir(dir, name string) error
	Remove(path string) error
}

// osFS is the default FileSystem, backed by the host operating system. It is a
// thin pass-through to the package-level os helpers, so the OS path stays
// byte-for-byte identical to operating on os.* directly.
type osFS struct{}

func (osFS) Open(name string) (fs.File, error)          { return os.Open(name) }
func (osFS) ReadDir(name string) ([]fs.DirEntry, error) { return os.ReadDir(name) }
func (osFS) Parent(p string) string                     { return ParentDir(p) }
func (osFS) Base(p string) string                       { return filepath.Base(p) }
func (osFS) Join(dir, name string) string               { return filepath.Join(dir, name) }
func (osFS) Display(p string) string                    { return p }
func (osFS) CreateFile(dir, name string) error          { return CreateFile(dir, name) }
func (osFS) CreateDir(dir, name string) error           { return CreateDir(dir, name) }
func (osFS) Remove(p string) error                      { return DeletePath(p) }

func (osFS) Root() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return dir
	}
	return abs
}

// fsAdapter wraps a standard io/fs.FS so it satisfies FileSystem. It operates
// in fs path space (slash-separated, relative to the FS root, "." for the
// root) and presents a clean breadcrumb so navigation feels identical to the
// OS picker. It is read-only, since io/fs has no write interface.
type fsAdapter struct {
	fsys  fs.FS
	label string // breadcrumb prefix for the root, e.g. "/"
}

func (a fsAdapter) Open(name string) (fs.File, error) {
	return a.fsys.Open(cleanFSPath(name))
}

func (a fsAdapter) ReadDir(name string) ([]fs.DirEntry, error) {
	return fs.ReadDir(a.fsys, cleanFSPath(name))
}

func (a fsAdapter) Parent(p string) string {
	return path.Dir(cleanFSPath(p))
}

func (a fsAdapter) Base(p string) string {
	return path.Base(cleanFSPath(p))
}

func (a fsAdapter) Join(dir, name string) string {
	return path.Join(cleanFSPath(dir), name)
}

func (a fsAdapter) Display(p string) string {
	p = cleanFSPath(p)
	if p == "." {
		return a.label
	}
	return strings.TrimRight(a.label, "/") + "/" + p
}

func (fsAdapter) Root() string { return "." }

// cleanFSPath normalizes a path into a valid io/fs path: rooted-relative,
// slash-separated, with "." for the root.
func cleanFSPath(p string) string {
	p = strings.TrimLeft(p, "/")
	if p == "" {
		return "."
	}
	return path.Clean(p)
}

// buildEntries turns raw directory entries into sorted FileEntry values,
// applying the hidden-file, glob, and extension filters. Paths are produced via
// the filesystem's Join so the result is correct for any backend. Directories
// sort before files; each group is sorted case-insensitively by name.
func buildEntries(fsys FileSystem, dir string, des []fs.DirEntry, showHidden bool, filters, extensions []string) []FileEntry {
	var result []FileEntry
	for _, e := range des {
		name := e.Name()

		if !showHidden && isHidden(name) {
			continue
		}

		info, err := e.Info()
		if err != nil {
			continue
		}

		entry := FileEntry{
			Name:     name,
			Path:     fsys.Join(dir, name),
			IsDir:    e.IsDir(),
			IsHidden: isHidden(name),
			Size:     info.Size(),
			Mode:     info.Mode(),
		}

		if !entry.IsDir && !passesFilters(name, filters, extensions) {
			continue
		}

		result = append(result, entry)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].IsDir != result[j].IsDir {
			return result[i].IsDir
		}
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	return result
}
