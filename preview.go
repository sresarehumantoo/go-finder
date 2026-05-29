package finder

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

const (
	// previewMaxBytes caps how much of a file is read for the preview.
	previewMaxBytes = 64 * 1024
	// previewMinTotalWidth is the minimum terminal width at which the preview
	// pane is shown; below this the picker stays single-column.
	previewMinTotalWidth = 80
	// previewMinPaneWidth is the minimum content width of the preview pane.
	previewMinPaneWidth = 24
	// listMinWidth is the minimum width reserved for the file list when the
	// preview pane is visible.
	listMinWidth = 30
)

// layout returns the content width of the file list, and the content width and
// height of the preview pane. When the preview is disabled or the terminal is
// too narrow, prevW and prevH are zero and listW spans the full width.
func (m Model) layout() (listW, prevW, prevH int) {
	if !m.options.Preview || m.width < previewMinTotalWidth {
		return m.width, 0, 0
	}
	const overhead = 2 // PreviewBorder left border + left padding

	prevW = m.width * 2 / 5
	if prevW < previewMinPaneWidth {
		prevW = previewMinPaneWidth
	}
	if m.width-prevW-overhead < listMinWidth {
		prevW = m.width - overhead - listMinWidth
	}
	if prevW < previewMinPaneWidth {
		return m.width, 0, 0
	}
	return m.width - prevW - overhead, prevW, m.visibleRows()
}

// refreshPreview recomputes the cached preview for the highlighted entry. It is
// called whenever the selection, directory, or terminal size changes.
func (m *Model) refreshPreview() {
	m.previewPath = ""
	m.previewLines = nil

	if !m.options.Preview {
		return
	}
	_, prevW, prevH := m.layout()
	if prevW <= 0 || prevH <= 0 {
		return
	}
	if m.cursor < 0 || m.cursor >= len(m.entries) {
		return
	}

	e := m.entries[m.cursor]
	m.previewPath = e.Path

	var text string
	if m.options.PreviewFunc != nil {
		text = m.options.PreviewFunc(e, prevW, prevH)
	} else {
		text = m.buildPreview(e, prevW, prevH)
	}
	m.previewLines = clipLines(text, prevW, prevH)
}

// buildPreview produces the default preview text for an entry: a metadata
// header followed by a directory listing or a file's head.
func (m Model) buildPreview(e FileEntry, _, height int) string {
	var b strings.Builder

	name := e.Name
	if e.IsDir {
		name += "/"
	}
	fmt.Fprintf(&b, "%s\n%s  %s\n\n", name, e.Mode.String(), FormatSize(e.Size))

	if e.IsDir {
		raw, err := m.fsys.ReadDir(e.Path)
		if err != nil {
			fmt.Fprintf(&b, "(cannot read: %v)", err)
			return b.String()
		}
		children := buildEntries(m.fsys, e.Path, raw, m.options.ShowHidden, nil)
		limit := height - 3
		if limit < 1 {
			limit = 1
		}
		for i, ce := range children {
			if i >= limit {
				fmt.Fprintf(&b, "… +%d more", len(children)-i)
				break
			}
			cn := ce.Name
			if ce.IsDir {
				cn += "/"
			}
			b.WriteString(cn)
			b.WriteString("\n")
		}
		return b.String()
	}

	f, err := m.fsys.Open(e.Path)
	if err != nil {
		fmt.Fprintf(&b, "(cannot preview: %v)", err)
		return b.String()
	}
	defer func() { _ = f.Close() }()

	data, err := io.ReadAll(io.LimitReader(f, previewMaxBytes))
	if err != nil {
		fmt.Fprintf(&b, "(cannot preview: %v)", err)
		return b.String()
	}
	if isBinary(data) {
		b.WriteString("(binary file)")
		return b.String()
	}
	b.Write(data)
	return b.String()
}

// isBinary reports whether data looks like a binary file. A NUL byte is a
// strong, truncation-safe signal (unlike a UTF-8 validity check, which a byte
// cap could split mid-rune).
func isBinary(data []byte) bool {
	return bytes.IndexByte(data, 0) != -1
}

// clipLines splits text into at most height lines, expands tabs, and clips each
// line to width cells.
func clipLines(text string, width, height int) []string {
	expanded := strings.ReplaceAll(text, "\t", "    ")
	raw := strings.Split(strings.TrimRight(expanded, "\n"), "\n")
	lines := make([]string, 0, height)
	for _, ln := range raw {
		if len(lines) >= height {
			break
		}
		lines = append(lines, clipWidth(ln, width))
	}
	return lines
}

// clipWidth truncates s to width runes, appending an ellipsis when shortened.
func clipWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}
	return string(r[:width-1]) + "…"
}
