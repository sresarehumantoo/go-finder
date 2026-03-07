package finder

import "github.com/charmbracelet/lipgloss"

// Styles holds all visual styles for the picker UI.
type Styles struct {
	Title        lipgloss.Style
	Path         lipgloss.Style
	Cursor       lipgloss.Style
	Selected     lipgloss.Style
	Directory    lipgloss.Style
	HiddenDir    lipgloss.Style
	File         lipgloss.Style
	HiddenFile   lipgloss.Style
	FileSize     lipgloss.Style
	Permission   lipgloss.Style
	StatusBar    lipgloss.Style
	SearchPrompt lipgloss.Style
	SearchText   lipgloss.Style
	Help         lipgloss.Style
	HelpKey      lipgloss.Style
	HelpDesc     lipgloss.Style
	HelpSep      lipgloss.Style
	Border       lipgloss.Style
}

// DefaultStyles returns the default color scheme and styling.
func DefaultStyles() Styles {
	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			PaddingBottom(1),

		Path: lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			PaddingBottom(1),

		Cursor: lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true),

		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("120")).
			Bold(true),

		Directory: lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true),

		HiddenDir: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true),

		File: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),

		HiddenFile: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true),

		FileSize: lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Width(10).
			Align(lipgloss.Right),

		Permission: lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")),

		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			PaddingTop(1),

		SearchPrompt: lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true),

		SearchText: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),

		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),

		HelpKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true),

		HelpDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")),

		HelpSep: lipgloss.NewStyle().
			Foreground(lipgloss.Color("237")),

		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2),
	}
}
