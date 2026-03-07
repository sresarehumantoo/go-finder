package finder

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// FileEntry represents a single file or directory in a listing.
type FileEntry struct {
	Name     string
	Path     string
	IsDir    bool
	IsHidden bool
	Size     int64
	Mode     os.FileMode
}

// ReadDir reads the contents of a directory and returns a sorted list of
// FileEntry values. Directories are sorted before files, and each group is
// sorted alphabetically (case-insensitive).
func ReadDir(dir string, showHidden bool, filters []string) ([]FileEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var result []FileEntry
	for _, e := range entries {
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
			Path:     filepath.Join(dir, name),
			IsDir:    e.IsDir(),
			IsHidden: isHidden(name),
			Size:     info.Size(),
			Mode:     info.Mode(),
		}

		if !entry.IsDir && len(filters) > 0 {
			if !matchesAnyFilter(name, filters) {
				continue
			}
		}

		result = append(result, entry)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].IsDir != result[j].IsDir {
			return result[i].IsDir
		}
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	return result, nil
}

// ParentDir returns the parent directory of the given path.
func ParentDir(dir string) string {
	parent := filepath.Dir(dir)
	if parent == dir {
		return dir
	}
	return parent
}

// NormalizePath converts a path to the OS-native format and resolves it
// to an absolute path.
func NormalizePath(p string) string {
	p = expandHome(p)
	abs, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	return filepath.Clean(abs)
}

// IsWSL returns true if running inside Windows Subsystem for Linux.
func IsWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	lower := strings.ToLower(string(data))
	return strings.Contains(lower, "microsoft") || strings.Contains(lower, "wsl")
}

// ToWindowsPath converts a WSL path like /mnt/c/Users to C:\Users.
// If the path is not a WSL mount path, it is returned unchanged.
func ToWindowsPath(p string) string {
	if !strings.HasPrefix(p, "/mnt/") || len(p) < 6 {
		return p
	}
	drive := strings.ToUpper(string(p[5]))
	rest := ""
	if len(p) > 6 {
		rest = strings.ReplaceAll(p[6:], "/", `\`)
	}
	return drive + ":" + rest
}

// ToWSLPath converts a Windows path like C:\Users to /mnt/c/Users.
// If the path is not a Windows-style path, it is returned unchanged.
func ToWSLPath(p string) string {
	if len(p) < 2 || p[1] != ':' {
		return p
	}
	drive := strings.ToLower(string(p[0]))
	rest := ""
	if len(p) > 2 {
		rest = strings.ReplaceAll(p[2:], `\`, "/")
	}
	return "/mnt/" + drive + rest
}

// FormatSize returns a human-readable file size string.
func FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return formatFloat(float64(bytes)/float64(GB)) + " GB"
	case bytes >= MB:
		return formatFloat(float64(bytes)/float64(MB)) + " MB"
	case bytes >= KB:
		return formatFloat(float64(bytes)/float64(KB)) + " KB"
	default:
		return formatInt(bytes) + " B"
	}
}

func formatFloat(f float64) string {
	return strings.TrimRight(strings.TrimRight(stringFromFloat(f), "0"), ".")
}

func stringFromFloat(f float64) string {
	whole := int64(f)
	frac := int64((f - float64(whole)) * 10)
	if frac < 0 {
		frac = -frac
	}
	return formatInt(whole) + "." + formatInt(frac)
}

func formatInt(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// CreateFile creates an empty file at the given path.
// Returns an error if the file already exists, the name is invalid, or it cannot be created.
func CreateFile(dir, name string) error {
	if err := validateName(name); err != nil {
		return err
	}
	p := filepath.Join(dir, name)
	f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	return f.Close()
}

// CreateDir creates a new directory at the given path.
// Returns an error if the directory already exists, the name is invalid, or it cannot be created.
func CreateDir(dir, name string) error {
	if err := validateName(name); err != nil {
		return err
	}
	p := filepath.Join(dir, name)
	return os.Mkdir(p, 0755)
}

// DeletePath removes a file or directory (and all contents) at the given path.
// Refuses to delete filesystem roots or paths that resolve to /.
func DeletePath(p string) error {
	abs, err := filepath.Abs(p)
	if err != nil {
		return fmt.Errorf("cannot resolve path: %w", err)
	}
	if abs == "/" {
		return errors.New("refusing to delete filesystem root")
	}
	home, _ := os.UserHomeDir()
	if home != "" && abs == filepath.Clean(home) {
		return errors.New("refusing to delete home directory")
	}
	return os.RemoveAll(abs)
}

// validateName checks that a file or directory name is safe.
func validateName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	if strings.Contains(name, "..") {
		return errors.New("name cannot contain '..'")
	}
	if strings.ContainsAny(name, "/\\") {
		return errors.New("name cannot contain path separators")
	}
	return nil
}

// ResolvePath evaluates symlinks and returns the real absolute path.
// If the path cannot be resolved, it falls back to filepath.Abs.
func ResolvePath(p string) string {
	p = expandHome(p)
	resolved, err := filepath.EvalSymlinks(p)
	if err != nil {
		abs, err2 := filepath.Abs(p)
		if err2 != nil {
			return p
		}
		return abs
	}
	abs, err := filepath.Abs(resolved)
	if err != nil {
		return resolved
	}
	return abs
}

// isHidden checks whether a file name represents a hidden file.
func isHidden(name string) bool {
	return strings.HasPrefix(name, ".")
}

// matchesAnyFilter checks if the file name matches at least one glob pattern.
func matchesAnyFilter(name string, filters []string) bool {
	for _, pattern := range filters {
		matched, err := filepath.Match(pattern, name)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// expandHome expands a leading ~ to the user's home directory.
func expandHome(p string) string {
	if !strings.HasPrefix(p, "~") {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}
	return filepath.Join(home, p[1:])
}
