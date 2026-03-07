package finder_test

import (
	"os"
	"path/filepath"
	"testing"

	finder "github.com/SREsAreHumanToo/go-finder"
)

func TestReadDir(t *testing.T) {
	// Create a temp directory with known files.
	dir := t.TempDir()
	createFile(t, dir, "alpha.go", "package alpha")
	createFile(t, dir, "beta.txt", "hello")
	createFile(t, dir, ".hidden", "secret")
	createDir(t, dir, "subdir")

	t.Run("no hidden files", func(t *testing.T) {
		entries, err := finder.ReadDir(dir, false, nil)
		if err != nil {
			t.Fatalf("ReadDir error: %v", err)
		}

		names := entryNames(entries)
		assertContains(t, names, "subdir")
		assertContains(t, names, "alpha.go")
		assertContains(t, names, "beta.txt")
		assertNotContains(t, names, ".hidden")
	})

	t.Run("with hidden files", func(t *testing.T) {
		entries, err := finder.ReadDir(dir, true, nil)
		if err != nil {
			t.Fatalf("ReadDir error: %v", err)
		}

		names := entryNames(entries)
		assertContains(t, names, ".hidden")
	})

	t.Run("with filter", func(t *testing.T) {
		entries, err := finder.ReadDir(dir, false, []string{"*.go"})
		if err != nil {
			t.Fatalf("ReadDir error: %v", err)
		}

		names := entryNames(entries)
		assertContains(t, names, "alpha.go")
		assertNotContains(t, names, "beta.txt")
		// Directories should always be included.
		assertContains(t, names, "subdir")
	})

	t.Run("directories sorted first", func(t *testing.T) {
		entries, err := finder.ReadDir(dir, false, nil)
		if err != nil {
			t.Fatalf("ReadDir error: %v", err)
		}
		if len(entries) == 0 {
			t.Fatal("expected entries, got none")
		}
		if !entries[0].IsDir {
			t.Errorf("expected first entry to be a directory, got %s", entries[0].Name)
		}
	})
}

func TestParentDir(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/home/user/projects", "/home/user"},
		{"/home", "/"},
		{"/", "/"},
	}

	for _, tt := range tests {
		got := finder.ParentDir(tt.input)
		if got != tt.expected {
			t.Errorf("ParentDir(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestNormalizePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}

	result := finder.NormalizePath("~/testpath")
	expected := filepath.Join(home, "testpath")
	if result != expected {
		t.Errorf("NormalizePath(~/testpath) = %q, want %q", result, expected)
	}
}

func TestToWindowsPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/mnt/c/Users/test", `C:\Users\test`},
		{"/mnt/d/data", `D:\data`},
		{"/home/user", "/home/user"},
		{"/mnt/c", `C:`},
	}

	for _, tt := range tests {
		got := finder.ToWindowsPath(tt.input)
		if got != tt.expected {
			t.Errorf("ToWindowsPath(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestToWSLPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`C:\Users\test`, "/mnt/c/Users/test"},
		{`D:\data`, "/mnt/d/data"},
		{"/home/user", "/home/user"},
	}

	for _, tt := range tests {
		got := finder.ToWSLPath(tt.input)
		if got != tt.expected {
			t.Errorf("ToWSLPath(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1 KB"},
		{1536, "1.5 KB"},
		{1048576, "1 MB"},
		{1073741824, "1 GB"},
	}

	for _, tt := range tests {
		got := finder.FormatSize(tt.bytes)
		if got != tt.expected {
			t.Errorf("FormatSize(%d) = %q, want %q", tt.bytes, got, tt.expected)
		}
	}
}

func TestCreateFilePathTraversal(t *testing.T) {
	dir := t.TempDir()

	// Should reject names with ".."
	err := finder.CreateFile(dir, "../evil.txt")
	if err == nil {
		t.Error("expected error for path traversal name '..'")
	}

	// Should reject names with path separators
	err = finder.CreateFile(dir, "sub/file.txt")
	if err == nil {
		t.Error("expected error for name with path separator")
	}

	// Should reject empty names
	err = finder.CreateFile(dir, "")
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestCreateDirPathTraversal(t *testing.T) {
	dir := t.TempDir()

	err := finder.CreateDir(dir, "../../evil")
	if err == nil {
		t.Error("expected error for path traversal in dir name")
	}

	err = finder.CreateDir(dir, "")
	if err == nil {
		t.Error("expected error for empty dir name")
	}
}

func TestDeletePathRefusesRoot(t *testing.T) {
	err := finder.DeletePath("/")
	if err == nil {
		t.Error("expected error when attempting to delete /")
	}
}

func TestDeletePathRefusesHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}
	err = finder.DeletePath(home)
	if err == nil {
		t.Error("expected error when attempting to delete home directory")
	}
}

func TestModeString(t *testing.T) {
	tests := []struct {
		mode     finder.Mode
		expected string
	}{
		{finder.ModeFile, "file"},
		{finder.ModeFolder, "folder"},
		{finder.ModeAny, "any"},
		{finder.ModeMultiple, "multiple"},
	}
	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.expected {
			t.Errorf("Mode(%d).String() = %q, want %q", tt.mode, got, tt.expected)
		}
	}
}

// Helpers.

func createFile(t *testing.T, dir, name, content string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create file %s: %v", name, err)
	}
}

func createDir(t *testing.T, parent, name string) {
	t.Helper()
	err := os.Mkdir(filepath.Join(parent, name), 0755)
	if err != nil {
		t.Fatalf("failed to create dir %s: %v", name, err)
	}
}

func entryNames(entries []finder.FileEntry) []string {
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name
	}
	return names
}

func assertContains(t *testing.T, list []string, item string) {
	t.Helper()
	for _, v := range list {
		if v == item {
			return
		}
	}
	t.Errorf("expected %v to contain %q", list, item)
}

func assertNotContains(t *testing.T, list []string, item string) {
	t.Helper()
	for _, v := range list {
		if v == item {
			t.Errorf("expected %v to NOT contain %q", list, item)
			return
		}
	}
}
