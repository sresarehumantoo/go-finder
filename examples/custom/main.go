// Custom example shows how to override keybindings and styles.
//
// Usage:
//
//	go run ./examples/custom
package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"

	finder "github.com/SREsAreHumanToo/go-finder"
)

func main() {
	// Custom keybindings: use 'x' to quit instead of 'q'.
	km := finder.DefaultKeyMap()
	km.Cancel = key.NewBinding(
		key.WithKeys("x", "ctrl+c"),
		key.WithHelp("x/ctrl+c", "quit"),
	)

	// Custom styles: green directories, yellow cursor.
	s := finder.DefaultStyles()
	s.Directory = lipgloss.NewStyle().
		Foreground(lipgloss.Color("82")).
		Bold(true)
	s.Cursor = lipgloss.NewStyle().
		Foreground(lipgloss.Color("220")).
		Bold(true)

	path, err := finder.PickFile(
		finder.WithTitle("Custom styled picker"),
		finder.WithKeyMap(km),
		finder.WithStyles(s),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Selected:", path)
}
