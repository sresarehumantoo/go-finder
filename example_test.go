package finder_test

import (
	"fmt"
	"testing/fstest"

	finder "github.com/SREsAreHumanToo/go-finder"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// Pick a single Go source file from the user's projects directory.
func ExamplePickFile() {
	path, err := finder.PickFile(
		finder.WithStartDir("~/projects"),
		finder.WithFilter("*.go"),
	)
	if err != nil {
		// finder.ErrCancelled is returned if the user exits without selecting.
		return
	}
	fmt.Println("selected:", path)
}

// Restrict the picker to a set of document extensions. Matching is
// case-insensitive, so "Report.PDF" is shown by WithExtensions("pdf").
func ExampleWithExtensions() {
	path, err := finder.PickFile(
		finder.WithExtensions("pdf", "docx", "doc"),
	)
	if err != nil {
		return
	}
	fmt.Println("document:", path)
}

// Pick a folder, using the s key inside the picker to confirm the current
// directory.
func ExamplePickFolder() {
	dir, err := finder.PickFolder()
	if err != nil {
		return
	}
	fmt.Println("folder:", dir)
}

// Pick multiple files matching a filter, with hidden files visible. Selections
// persist across directory navigation.
func ExamplePickMultiple() {
	paths, err := finder.PickMultiple(
		finder.WithFilter("*.log", "*.txt"),
		finder.WithHidden(true),
	)
	if err != nil {
		return
	}
	for _, p := range paths {
		fmt.Println(p)
	}
}

// Rebind the cancel key from q to x.
func ExampleWithKeyMap() {
	km := finder.DefaultKeyMap()
	km.Cancel = key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "exit"),
	)
	_, _ = finder.PickFile(finder.WithKeyMap(km))
}

// Browse an in-memory filesystem instead of the host OS. Any io/fs.FS works,
// including embed.FS and fstest.MapFS. Custom filesystems are read-only, so
// interactive create/delete is disabled for them.
func ExampleWithFS() {
	mem := fstest.MapFS{
		"docs/readme.md": {Data: []byte("# hello")},
		"main.go":        {Data: []byte("package main")},
	}
	path, err := finder.PickFile(finder.WithFS(mem))
	if err != nil {
		return
	}
	fmt.Println("selected:", path)
}

// Embed the picker as a sub-model in your own Bubble Tea program. New returns a
// model in embedded mode: forward messages to its Update, render with View, and
// watch for finder.DoneMsg to read the result. The picker never quits your
// program itself.
func ExampleNew() {
	picker := finder.New(
		finder.WithMode(finder.ModeFile),
		finder.WithTitle("Pick a file"),
	)

	// In your parent model's Update, forward messages and handle completion:
	//
	//	func (m parent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	//		if done, ok := msg.(finder.DoneMsg); ok {
	//			if done.State == finder.StateSelected {
	//				m.result = done.Paths
	//			}
	//			return m, nil
	//		}
	//		updated, cmd := m.picker.Update(msg)
	//		m.picker = updated.(finder.Model)
	//		return m, cmd
	//	}
	_ = picker.Init()
}

// Show a preview pane beside the file list. The pane previews the highlighted
// entry (file head, directory listing, or metadata) and is hidden automatically
// on narrow terminals.
func ExampleWithPreview() {
	path, err := finder.PickFile(finder.WithPreview(true))
	if err != nil {
		return
	}
	fmt.Println("selected:", path)
}

// Supply a custom preview renderer. The function receives the highlighted entry
// and the pane dimensions, and returns the text to display.
func ExampleWithPreviewFunc() {
	_, _ = finder.PickFile(finder.WithPreviewFunc(
		func(e finder.FileEntry, width, height int) string {
			return fmt.Sprintf("%s\n%s", e.Name, finder.FormatSize(e.Size))
		},
	))
}

// Override visual styles for directory entries.
func ExampleWithStyles() {
	s := finder.DefaultStyles()
	s.Directory = lipgloss.NewStyle().
		Foreground(lipgloss.Color("82")).
		Bold(true)
	_, _ = finder.PickFile(finder.WithStyles(s))
}

func ExampleFormatSize() {
	fmt.Println(finder.FormatSize(0))
	fmt.Println(finder.FormatSize(1536))
	fmt.Println(finder.FormatSize(1048576))
	fmt.Println(finder.FormatSize(1073741824))
	// Output:
	// 0 B
	// 1.5 KB
	// 1 MB
	// 1 GB
}

func ExampleToWindowsPath() {
	fmt.Println(finder.ToWindowsPath("/mnt/c/Users/test"))
	fmt.Println(finder.ToWindowsPath("/mnt/d/data"))
	fmt.Println(finder.ToWindowsPath("/home/user"))
	// Output:
	// C:\Users\test
	// D:\data
	// /home/user
}

func ExampleToWSLPath() {
	fmt.Println(finder.ToWSLPath(`C:\Users\test`))
	fmt.Println(finder.ToWSLPath(`D:\data`))
	fmt.Println(finder.ToWSLPath("/home/user"))
	// Output:
	// /mnt/c/Users/test
	// /mnt/d/data
	// /home/user
}

// ParentDir returns the parent of a path, or the path itself if it has no
// parent (e.g. the filesystem root). The output uses the OS-native path
// separator, so on Windows /home/user becomes \home — for portable usage,
// pass paths through filepath.FromSlash first.
func ExampleParentDir() {
	parent := finder.ParentDir("/home/user/projects")
	fmt.Println(parent)
}

func ExampleMode_String() {
	fmt.Println(finder.ModeFile)
	fmt.Println(finder.ModeFolder)
	fmt.Println(finder.ModeAny)
	fmt.Println(finder.ModeMultiple)
	// Output:
	// file
	// folder
	// any
	// multiple
}
