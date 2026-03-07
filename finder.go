package finder

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// ErrCancelled is returned when the user cancels the picker without selecting.
var ErrCancelled = errors.New("picker cancelled")

// PickFile opens an interactive file picker and returns the selected file path.
// Returns ErrCancelled if the user exits without selecting.
func PickFile(opts ...Option) (string, error) {
	o := DefaultOptions()
	o.Mode = ModeFile
	o.Title = "Select a file"
	for _, fn := range opts {
		fn(&o)
	}

	result, err := runPicker(o)
	if err != nil {
		return "", err
	}
	if len(result) == 0 {
		return "", ErrCancelled
	}
	return result[0], nil
}

// PickFolder opens an interactive folder picker and returns the selected directory path.
// Returns ErrCancelled if the user exits without selecting.
func PickFolder(opts ...Option) (string, error) {
	o := DefaultOptions()
	o.Mode = ModeFolder
	o.Title = "Select a folder"
	for _, fn := range opts {
		fn(&o)
	}

	result, err := runPicker(o)
	if err != nil {
		return "", err
	}
	if len(result) == 0 {
		return "", ErrCancelled
	}
	return result[0], nil
}

// PickAny opens an interactive picker that allows selecting either a file or
// a folder. Press enter to select the highlighted item, right/l to navigate
// into directories, or s to select the current directory.
// Returns ErrCancelled if the user exits without selecting.
func PickAny(opts ...Option) (string, error) {
	o := DefaultOptions()
	o.Mode = ModeAny
	o.Title = "Select a file or folder"
	for _, fn := range opts {
		fn(&o)
	}

	result, err := runPicker(o)
	if err != nil {
		return "", err
	}
	if len(result) == 0 {
		return "", ErrCancelled
	}
	return result[0], nil
}

// PickMultiple opens an interactive multi-select file picker and returns
// all selected paths. Returns ErrCancelled if the user exits without selecting.
func PickMultiple(opts ...Option) ([]string, error) {
	o := DefaultOptions()
	o.Mode = ModeMultiple
	o.Title = "Select files (space to toggle, enter to confirm)"
	for _, fn := range opts {
		fn(&o)
	}

	result, err := runPicker(o)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, ErrCancelled
	}
	return result, nil
}

func runPicker(opts Options) ([]string, error) {
	m := NewModel(opts)

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("picker error: %w", err)
	}

	fm := finalModel.(Model)
	if fm.Err() != nil {
		return nil, fm.Err()
	}

	return fm.SelectedPaths(), nil
}
