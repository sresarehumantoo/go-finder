// Embedded example shows how to embed the picker as a sub-model inside your
// own Bubble Tea program, rather than using the standalone Pick* API.
//
// The parent owns the tea.Program. It forwards messages to the picker, watches
// for finder.DoneMsg to learn the result, and then switches to its own view.
//
// Usage:
//
//	go run ./examples/embedded
package main

import (
	"fmt"
	"os"

	finder "github.com/rummage-dev/rummage"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	picker   finder.Model
	picking  bool
	selected string
	canceled bool
}

func initialModel() model {
	return model{
		picker: finder.New(
			finder.WithMode(finder.ModeFile),
			finder.WithTitle("Pick a file (embedded)"),
			finder.WithPreview(true),
		),
		picking: true,
	}
}

func (m model) Init() tea.Cmd {
	return m.picker.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Let the parent always quit on ctrl+c, even while the picker is active.
	if key, ok := msg.(tea.KeyMsg); ok && key.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	// React to the picker finishing.
	if done, ok := msg.(finder.DoneMsg); ok {
		m.picking = false
		switch done.State {
		case finder.StateSelected:
			if len(done.Paths) > 0 {
				m.selected = done.Paths[0]
			}
		case finder.StateCancelled:
			m.canceled = true
		}
		return m, tea.Quit
	}

	if m.picking {
		updated, cmd := m.picker.Update(msg)
		m.picker = updated.(finder.Model)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.picking {
		return m.picker.View()
	}
	return ""
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	res := final.(model)
	switch {
	case res.canceled:
		fmt.Println("Cancelled.")
	case res.selected != "":
		fmt.Println("Selected:", res.selected)
	default:
		fmt.Println("Nothing selected.")
	}
}
