package finder

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keybindings for the file picker.
type KeyMap struct {
	Up        key.Binding
	Down      key.Binding
	PageUp    key.Binding
	PageDown  key.Binding
	Home      key.Binding
	End       key.Binding
	Navigate  key.Binding
	Back      key.Binding
	Select    key.Binding
	SelectDir key.Binding
	Toggle    key.Binding
	ToggleAll key.Binding
	Hidden    key.Binding
	Cancel    key.Binding
	Search    key.Binding
	NewFile   key.Binding
	NewFolder key.Binding
	Delete    key.Binding
}

// DefaultKeyMap returns the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("up/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("down/j", "move down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("pgdn", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("home/g", "go to top"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("end/G", "go to bottom"),
		),
		Navigate: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("right/l", "open directory"),
		),
		Back: key.NewBinding(
			key.WithKeys("backspace", "left", "h", "esc"),
			key.WithHelp("backspace/esc", "go back"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		SelectDir: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "select current directory"),
		),
		Toggle: key.NewBinding(
			key.WithKeys(" ", "tab"),
			key.WithHelp("space/tab", "toggle selection"),
		),
		ToggleAll: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "toggle all"),
		),
		Hidden: key.NewBinding(
			key.WithKeys("."),
			key.WithHelp(".", "toggle hidden"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "cancel"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		NewFile: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new file"),
		),
		NewFolder: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("N", "new folder"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d", "delete"),
			key.WithHelp("d/del", "delete"),
		),
	}
}
