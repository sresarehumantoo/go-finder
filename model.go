package finder

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

// Model is the bubbletea model for the file picker.
type Model struct {
	options  Options
	keys     KeyMap
	styles   Styles
	entries  []FileEntry
	cursor   int
	offset   int
	selected map[string]struct{}
	dir      string
	err      error
	quitting bool
	chosen   bool

	searching  bool
	searchTerm string
	allEntries []FileEntry

	inputMode inputModeType
	inputText string

	returnTo     string
	statusMsg    string
	hiddenForced bool

	width  int
	height int
}

// NewModel creates a new picker model with the given options.
func NewModel(opts Options) Model {
	w, h, _ := term.GetSize(0)
	if w == 0 {
		w = 80
	}
	if h == 0 {
		h = 24
	}

	startDir := opts.StartDir
	if opts.ExpandSymlinks {
		startDir = ResolvePath(startDir)
		opts.StartDir = startDir
	}

	keys := DefaultKeyMap()
	if opts.KeyMap != nil {
		keys = *opts.KeyMap
	}
	styles := DefaultStyles()
	if opts.Styles != nil {
		styles = *opts.Styles
	}

	m := Model{
		options:      opts,
		keys:         keys,
		styles:       styles,
		selected:     make(map[string]struct{}),
		dir:          startDir,
		hiddenForced: opts.ShowHidden,
		width:        w,
		height:       h,
	}

	return m
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return m.readDir()
}

// SelectedPath returns the single selected path after the picker closes.
// Returns empty string if nothing was selected.
func (m Model) SelectedPath() string {
	if !m.chosen {
		return ""
	}
	if m.options.Mode == ModeMultiple {
		for p := range m.selected {
			return p
		}
		return ""
	}
	if m.cursor >= 0 && m.cursor < len(m.entries) {
		return m.entries[m.cursor].Path
	}
	return ""
}

// SelectedPaths returns all selected paths (for multi-select mode).
// Paths are returned in sorted order for deterministic results.
func (m Model) SelectedPaths() []string {
	if !m.chosen {
		return nil
	}
	if m.options.Mode == ModeMultiple {
		paths := make([]string, 0, len(m.selected))
		for p := range m.selected {
			paths = append(paths, p)
		}
		sort.Strings(paths)
		return paths
	}
	if p := m.SelectedPath(); p != "" {
		return []string{p}
	}
	return nil
}

// Err returns any error that occurred during picker operation.
func (m Model) Err() error {
	return m.err
}

// dirReadMsg is sent after a directory read completes.
type dirReadMsg struct {
	entries []FileEntry
	dir     string
	err     error
}

func (m Model) readDir() tea.Cmd {
	dir := m.dir
	showHidden := m.options.ShowHidden
	filters := m.options.Filters
	return func() tea.Msg {
		entries, err := ReadDir(dir, showHidden, filters)
		return dirReadMsg{entries: entries, dir: dir, err: err}
	}
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case dirReadMsg:
		if msg.err != nil {
			m.err = msg.err
			m.entries = nil
			m.cursor = 0
			m.offset = 0
			return m, nil
		}
		m.entries = msg.entries
		m.dir = msg.dir
		m.cursor = 0
		m.offset = 0
		m.err = nil
		m.statusMsg = ""

		if m.returnTo != "" {
			for i, e := range m.entries {
				if e.Name == m.returnTo {
					m.cursor = i
					m.fixOffset()
					break
				}
			}
			m.returnTo = ""
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.searching {
		return m.handleSearchKey(msg)
	}
	if m.inputMode != inputNone {
		return m.handleInputKey(msg)
	}

	m.statusMsg = ""

	switch {
	case key.Matches(msg, m.keys.Cancel):
		m.quitting = true
		return m, tea.Quit

	case key.Matches(msg, m.keys.Up):
		m.cursorUp()

	case key.Matches(msg, m.keys.Down):
		m.cursorDown()

	case key.Matches(msg, m.keys.PageUp):
		m.pageUp()

	case key.Matches(msg, m.keys.PageDown):
		m.pageDown()

	case key.Matches(msg, m.keys.Home):
		m.cursor = 0
		m.offset = 0

	case key.Matches(msg, m.keys.End):
		m.cursor = max(0, len(m.entries)-1)
		m.fixOffset()

	case key.Matches(msg, m.keys.Hidden) && !m.hiddenForced:
		m.options.ShowHidden = !m.options.ShowHidden
		return m, m.readDir()

	case key.Matches(msg, m.keys.Search):
		m.searching = true
		m.searchTerm = ""
		m.allEntries = make([]FileEntry, len(m.entries))
		copy(m.allEntries, m.entries)

	case key.Matches(msg, m.keys.Back):
		parent := m.resolve(ParentDir(m.dir))
		if parent != m.dir {
			m.returnTo = filepath.Base(m.dir)
			m.dir = parent
			return m, m.readDir()
		}
		m.statusMsg = "At highest level — press q to quit"

	case key.Matches(msg, m.keys.Navigate):
		return m.handleNavigate()

	case key.Matches(msg, m.keys.SelectDir):
		return m.handleSelectDir()

	case key.Matches(msg, m.keys.NewFile) && m.options.Interactive && m.options.Mode != ModeFolder:
		m.inputMode = inputNewFile
		m.inputText = ""

	case key.Matches(msg, m.keys.NewFolder) && m.options.Interactive,
		key.Matches(msg, m.keys.NewFile) && m.options.Interactive && m.options.Mode == ModeFolder:
		m.inputMode = inputNewFolder
		m.inputText = ""

	case key.Matches(msg, m.keys.Delete) && m.options.Interactive:
		if len(m.entries) > 0 {
			m.inputMode = inputConfirmDelete
		}

	default:
		return m.handleSelectionKey(msg)
	}

	return m, nil
}

// handleNavigate handles right/l — always navigates into directories.
// In multi-select mode, selections are preserved across directories.
func (m Model) handleNavigate() (tea.Model, tea.Cmd) {
	if len(m.entries) == 0 {
		return m, nil
	}
	entry := m.entries[m.cursor]
	if entry.IsDir {
		m.dir = m.resolve(entry.Path)
		return m, m.readDir()
	}
	return m, nil
}

// handleSelectDir handles 's' — selects the current working directory.
func (m Model) handleSelectDir() (tea.Model, tea.Cmd) {
	if m.options.Mode == ModeFolder || m.options.Mode == ModeAny {
		m.chosen = true
		m.entries = []FileEntry{{
			Name:  ".",
			Path:  m.dir,
			IsDir: true,
		}}
		m.cursor = 0
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleSelectionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.entries) == 0 {
		return m, nil
	}

	entry := m.entries[m.cursor]

	switch {
	case key.Matches(msg, m.keys.Toggle) && m.options.Mode == ModeAny:
		m.chosen = true
		return m, tea.Quit

	case key.Matches(msg, m.keys.Toggle) && m.options.Mode == ModeMultiple:
		if _, ok := m.selected[entry.Path]; ok {
			delete(m.selected, entry.Path)
		} else {
			m.selected[entry.Path] = struct{}{}
		}
		m.cursorDown()

	case key.Matches(msg, m.keys.ToggleAll) && m.options.Mode == ModeMultiple:
		allSelected := true
		for _, e := range m.entries {
			if _, ok := m.selected[e.Path]; !ok {
				allSelected = false
				break
			}
		}
		if allSelected {
			for _, e := range m.entries {
				delete(m.selected, e.Path)
			}
		} else {
			for _, e := range m.entries {
				m.selected[e.Path] = struct{}{}
			}
		}

	case key.Matches(msg, m.keys.Select):
		switch m.options.Mode {
		case ModeFile:
			if entry.IsDir {
				m.dir = m.resolve(entry.Path)
				return m, m.readDir()
			}
			m.chosen = true
			return m, tea.Quit

		case ModeFolder:
			if entry.IsDir {
				m.chosen = true
				return m, tea.Quit
			}

		case ModeAny:
			if entry.IsDir {
				m.dir = m.resolve(entry.Path)
				return m, m.readDir()
			}
			m.chosen = true
			return m, tea.Quit

		case ModeMultiple:
			if entry.IsDir {
				m.dir = m.resolve(entry.Path)
				return m, m.readDir()
			}
			if len(m.selected) > 0 {
				m.chosen = true
				return m, tea.Quit
			}
			m.selected[entry.Path] = struct{}{}
			m.chosen = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.searching = false
		m.searchTerm = ""
		m.entries = m.allEntries
		m.allEntries = nil
		m.cursor = 0
		m.offset = 0
		return m, nil

	case tea.KeyEnter:
		m.searching = false
		m.searchTerm = ""
		m.allEntries = nil
		if m.cursor >= len(m.entries) {
			m.cursor = max(0, len(m.entries)-1)
		}
		m.offset = 0
		m.fixOffset()
		return m, nil

	case tea.KeyBackspace:
		if len(m.searchTerm) > 0 {
			runes := []rune(m.searchTerm)
			m.searchTerm = string(runes[:len(runes)-1])
		}
		m.filterEntries()
		return m, nil

	default:
		if msg.Type == tea.KeyRunes {
			m.searchTerm += string(msg.Runes)
			m.filterEntries()
		}
		return m, nil
	}
}

// handleInputKey processes keyboard input during interactive prompts
// (new file, new folder, confirm delete).
func (m Model) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.inputMode {
	case inputNewFile, inputNewFolder:
		return m.handleCreateInput(msg)
	case inputConfirmDelete:
		return m.handleDeleteConfirm(msg)
	}
	return m, nil
}

func (m Model) handleCreateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.inputMode = inputNone
		m.inputText = ""
		return m, nil

	case tea.KeyEnter:
		name := strings.TrimSpace(m.inputText)
		if name == "" {
			m.inputMode = inputNone
			m.inputText = ""
			return m, nil
		}

		var err error
		if m.inputMode == inputNewFile && strings.HasSuffix(name, "/") {
			name = strings.TrimRight(name, "/")
			if name == "" {
				m.inputMode = inputNone
				m.inputText = ""
				return m, nil
			}
			err = CreateDir(m.dir, name)
		} else if m.inputMode == inputNewFile {
			err = CreateFile(m.dir, name)
		} else {
			err = CreateDir(m.dir, name)
		}

		m.inputMode = inputNone
		m.inputText = ""

		if err != nil {
			m.statusMsg = "Error: " + err.Error()
			return m, nil
		}

		m.returnTo = name
		return m, m.readDir()

	case tea.KeyBackspace:
		if len(m.inputText) > 0 {
			runes := []rune(m.inputText)
			m.inputText = string(runes[:len(runes)-1])
		}
		return m, nil

	default:
		if msg.Type == tea.KeyRunes {
			m.inputText += string(msg.Runes)
		} else if msg.Type == tea.KeySpace {
			m.inputText += " "
		}
		return m, nil
	}
}

func (m Model) handleDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.inputMode = inputNone

	if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && (msg.Runes[0] == 'y' || msg.Runes[0] == 'Y') {
		if m.cursor >= 0 && m.cursor < len(m.entries) {
			entry := m.entries[m.cursor]
			err := DeletePath(entry.Path)
			if err != nil {
				m.statusMsg = "Error: " + err.Error()
				return m, nil
			}
			m.statusMsg = "Deleted: " + entry.Name
			if m.cursor >= len(m.entries)-1 && m.cursor > 0 {
				m.cursor--
			}
			return m, m.readDir()
		}
	}

	return m, nil
}

// filterEntries filters the visible entries based on the current search term.
func (m *Model) filterEntries() {
	if m.searchTerm == "" {
		m.entries = m.allEntries
		m.cursor = 0
		m.offset = 0
		return
	}
	term := strings.ToLower(m.searchTerm)
	var filtered []FileEntry
	for _, e := range m.allEntries {
		if strings.Contains(strings.ToLower(e.Name), term) {
			filtered = append(filtered, e)
		}
	}
	m.entries = filtered
	m.cursor = 0
	m.offset = 0
}

func (m *Model) cursorUp() {
	if m.cursor > 0 {
		m.cursor--
		if m.cursor < m.offset {
			m.offset = m.cursor
		}
	}
}

func (m *Model) cursorDown() {
	if m.cursor < len(m.entries)-1 {
		m.cursor++
		m.fixOffset()
	}
}

func (m *Model) pageUp() {
	visible := m.visibleRows()
	m.cursor -= visible
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.offset -= visible
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m *Model) pageDown() {
	if len(m.entries) == 0 {
		return
	}
	visible := m.visibleRows()
	m.cursor += visible
	if m.cursor >= len(m.entries) {
		m.cursor = len(m.entries) - 1
	}
	m.fixOffset()
}

func (m *Model) fixOffset() {
	visible := m.visibleRows()
	if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m Model) visibleRows() int {
	reserved := 8
	h := m.height - reserved
	if m.options.Height > 0 && m.options.Height < h {
		h = m.options.Height
	}
	if h < 3 {
		h = 3
	}
	return h
}

// resolve applies symlink resolution if ExpandSymlinks is enabled.
func (m Model) resolve(p string) string {
	if m.options.ExpandSymlinks {
		return ResolvePath(p)
	}
	return p
}

type inputModeType int

const (
	inputNone inputModeType = iota
	inputNewFile
	inputNewFolder
	inputConfirmDelete
)
