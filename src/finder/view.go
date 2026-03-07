package finder

import (
	"fmt"
	"strings"
)

// View implements tea.Model. It renders the full picker UI.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString(m.styles.Title.Render(m.options.Title))
	b.WriteString("\n")

	displayDir := m.dir
	maxPath := m.width - 4
	if maxPath > 0 && len(displayDir) > maxPath {
		displayDir = "…" + displayDir[len(displayDir)-maxPath+1:]
	}
	b.WriteString(m.styles.Path.Render(displayDir))
	b.WriteString("\n")

	if m.options.Mode == ModeFolder || m.options.Mode == ModeAny {
		b.WriteString(m.styles.Help.Render("  press 's' to select this directory"))
		b.WriteString("\n")
	}

	if m.err != nil {
		b.WriteString(m.styles.Cursor.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n")
		return b.String()
	}

	if len(m.entries) == 0 {
		b.WriteString(m.styles.Help.Render("  (empty directory)"))
		b.WriteString("\n")
	}

	visible := m.visibleRows()
	end := m.offset + visible
	if end > len(m.entries) {
		end = len(m.entries)
	}

	for i := m.offset; i < end; i++ {
		entry := m.entries[i]
		isCursor := i == m.cursor
		_, isSelected := m.selected[entry.Path]

		line := m.renderEntry(entry, isCursor, isSelected)
		b.WriteString(line)
		b.WriteString("\n")
	}

	var statusParts []string
	if len(m.entries) > visible {
		statusParts = append(statusParts, fmt.Sprintf("%d/%d", m.cursor+1, len(m.entries)))
	}
	if m.options.Mode == ModeMultiple && len(m.selected) > 0 {
		statusParts = append(statusParts, fmt.Sprintf("%d selected", len(m.selected)))
	}
	if len(statusParts) > 0 {
		b.WriteString(m.styles.StatusBar.Render("  " + strings.Join(statusParts, "  •  ")))
		b.WriteString("\n")
	}

	if m.searching {
		b.WriteString(m.styles.SearchPrompt.Render("/"))
		b.WriteString(m.styles.SearchText.Render(m.searchTerm))
		b.WriteString(m.styles.SearchPrompt.Render("_"))
		b.WriteString("\n")
	}

	switch m.inputMode {
	case inputNewFile:
		b.WriteString(m.styles.SearchPrompt.Render("  New file: "))
		b.WriteString(m.styles.SearchText.Render(m.inputText))
		b.WriteString(m.styles.SearchPrompt.Render("_"))
		b.WriteString("\n")
	case inputNewFolder:
		b.WriteString(m.styles.SearchPrompt.Render("  New folder: "))
		b.WriteString(m.styles.SearchText.Render(m.inputText))
		b.WriteString(m.styles.SearchPrompt.Render("_"))
		b.WriteString("\n")
	case inputConfirmDelete:
		if m.cursor >= 0 && m.cursor < len(m.entries) {
			name := m.entries[m.cursor].Name
			if m.entries[m.cursor].IsDir {
				name += "/"
			}
			b.WriteString(m.styles.Cursor.Render(
				fmt.Sprintf("  Delete %s? (y/n)", name),
			))
			b.WriteString("\n")
		}
	}

	if m.statusMsg != "" {
		b.WriteString(m.styles.Cursor.Render("  " + m.statusMsg))
		b.WriteString("\n")
	}

	b.WriteString(m.renderHelp())

	return b.String()
}

func (m Model) renderEntry(e FileEntry, isCursor, isSelected bool) string {
	var prefix string
	if isCursor {
		prefix = m.styles.Cursor.Render("> ")
	} else {
		prefix = "  "
	}

	var marker string
	markerWidth := 0
	if m.options.Mode == ModeMultiple {
		markerWidth = 4
		if isSelected {
			marker = m.styles.Selected.Render("[x] ")
		} else {
			marker = "[ ] "
		}
	}

	sizeWidth := 10
	overhead := 2 + markerWidth + 2 + sizeWidth
	maxName := m.width - overhead
	if maxName < 10 {
		maxName = 10
	}

	displayName := e.Name
	if e.IsDir {
		displayName += "/"
	}
	if len(displayName) > maxName {
		displayName = displayName[:maxName-1] + "…"
	}

	var name string
	if e.IsDir {
		if isSelected {
			name = m.styles.Selected.Render(displayName)
		} else if e.IsHidden {
			name = m.styles.HiddenDir.Render(displayName)
		} else {
			name = m.styles.Directory.Render(displayName)
		}
	} else {
		if isSelected {
			name = m.styles.Selected.Render(displayName)
		} else if e.IsHidden {
			name = m.styles.HiddenFile.Render(displayName)
		} else {
			name = m.styles.File.Render(displayName)
		}
	}

	var size string
	if !e.IsDir {
		size = m.styles.FileSize.Render(FormatSize(e.Size))
	} else {
		size = m.styles.FileSize.Render("")
	}

	return prefix + marker + name + "  " + size
}

// helpBinding is a key-description pair for the help bar.
type helpBinding struct {
	key  string
	desc string
}

func (m Model) renderHelp() string {
	var bindings []helpBinding

	switch m.options.Mode {
	case ModeFile:
		bindings = append(bindings,
			helpBinding{m.keys.Select.Help().Key, "open/select"},
		)
	case ModeFolder:
		bindings = append(bindings,
			helpBinding{m.keys.Select.Help().Key, "select dir"},
			helpBinding{m.keys.Navigate.Help().Key, "open dir"},
			helpBinding{m.keys.SelectDir.Help().Key, "select here"},
		)
	case ModeAny:
		bindings = append(bindings,
			helpBinding{m.keys.Select.Help().Key, "open/select"},
			helpBinding{m.keys.Toggle.Help().Key, "select item"},
			helpBinding{m.keys.SelectDir.Help().Key, "select here"},
		)
	case ModeMultiple:
		bindings = append(bindings,
			helpBinding{m.keys.Toggle.Help().Key, "toggle"},
			helpBinding{m.keys.ToggleAll.Help().Key, "toggle all"},
			helpBinding{m.keys.Navigate.Help().Key, "open dir"},
			helpBinding{m.keys.Select.Help().Key, "confirm"},
		)
	}

	bindings = append(bindings,
		helpBinding{m.keys.Back.Help().Key, m.keys.Back.Help().Desc},
		helpBinding{m.keys.Search.Help().Key, m.keys.Search.Help().Desc},
	)
	if !m.hiddenForced {
		bindings = append(bindings,
			helpBinding{m.keys.Hidden.Help().Key, m.keys.Hidden.Help().Desc},
		)
	}

	if m.options.Interactive {
		bindings = append(bindings,
			helpBinding{m.keys.NewFile.Help().Key, m.keys.NewFile.Help().Desc},
			helpBinding{m.keys.NewFolder.Help().Key, m.keys.NewFolder.Help().Desc},
			helpBinding{m.keys.Delete.Help().Key, m.keys.Delete.Help().Desc},
		)
	}

	bindings = append(bindings,
		helpBinding{m.keys.Cancel.Help().Key, "quit"},
	)

	sep := m.styles.HelpSep.Render(" | ")
	var rendered []string
	for _, b := range bindings {
		item := m.styles.HelpKey.Render(b.key) + " " + m.styles.HelpDesc.Render(b.desc)
		rendered = append(rendered, item)
	}

	maxWidth := m.width
	if maxWidth <= 0 {
		maxWidth = 80
	}

	sepWidth := 3
	var lines []string
	var currentLine string
	currentWidth := 0

	for i, item := range rendered {
		visibleWidth := len(bindings[i].key) + 1 + len(bindings[i].desc)

		needed := visibleWidth
		if currentWidth > 0 {
			needed += sepWidth
		}

		if currentWidth > 0 && currentWidth+needed > maxWidth {
			lines = append(lines, currentLine)
			currentLine = item
			currentWidth = visibleWidth
		} else {
			if currentWidth > 0 {
				currentLine += sep + item
				currentWidth += sepWidth + visibleWidth
			} else {
				currentLine = item
				currentWidth = visibleWidth
			}
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n")
}
