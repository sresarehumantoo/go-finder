package finder

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sahilm/fuzzy"
	"golang.org/x/term"
)

// State describes where the picker is in its lifecycle. It lets a parent
// program that embeds the picker (see New) detect when the user has finished.
type State int

const (
	// StateBrowsing means the user is still navigating and has not finished.
	StateBrowsing State = iota
	// StateSelected means the user confirmed a selection; read it with
	// SelectedPath or SelectedPaths.
	StateSelected
	// StateCancelled means the user dismissed the picker without selecting.
	StateCancelled
)

// DoneMsg is emitted (as a tea.Cmd result) when an embedded picker finishes,
// so a parent program can react in its own Update. It is not sent in the
// standalone Pick* API, which ends its own program instead.
type DoneMsg struct {
	State State
	Paths []string
}

// Model is the bubbletea model for the file picker.
type Model struct {
	options  Options
	keys     KeyMap
	styles   Styles
	fsys     FileSystem
	entries  []FileEntry
	cursor   int
	offset   int
	selected map[string]struct{}
	dir      string
	err      error
	quitting bool
	state    State

	// standalone is true when the model owns its tea.Program (the Pick*
	// one-liner API) and may therefore emit tea.Quit. When embedded in a
	// parent program it is false: completion is signalled via DoneMsg instead.
	standalone bool

	searching  bool
	searchTerm string
	allEntries []FileEntry
	// matchIdx maps an entry path to the byte offsets of its fuzzy-matched
	// characters, used to highlight matches while searching.
	matchIdx map[string][]int

	inputMode inputModeType
	inputText string

	returnTo     string
	statusMsg    string
	hiddenForced bool

	previewPath  string
	previewLines []string

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

	var fsys FileSystem = osFS{}
	if opts.FS != nil {
		fsys = opts.FS
	}
	_, isOS := fsys.(osFS)

	startDir := opts.StartDir
	if isOS {
		if startDir == "" {
			startDir = fsys.Root()
		}
		startDir = expandHome(startDir)
		if abs, err := filepath.Abs(startDir); err == nil {
			startDir = abs
		}
		if opts.ExpandSymlinks {
			startDir = ResolvePath(startDir)
		}
	} else {
		// Symlink expansion is OS-specific and meaningless for a custom FS.
		opts.ExpandSymlinks = false
		// DefaultOptions seeds StartDir with an absolute OS path, which has no
		// meaning here; fall back to the filesystem root unless the caller gave
		// a relative in-FS path.
		if startDir == "" || filepath.IsAbs(startDir) {
			startDir = fsys.Root()
		} else {
			startDir = cleanFSPath(startDir)
		}
	}
	opts.StartDir = startDir

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
		fsys:         fsys,
		selected:     make(map[string]struct{}),
		dir:          startDir,
		hiddenForced: opts.ShowHidden,
		standalone:   !opts.Embedded,
		width:        w,
		height:       h,
	}

	return m
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return m.readDir()
}

// State returns the picker's lifecycle state. Embedded parents use it (or watch
// for DoneMsg) to detect when the user has finished.
func (m Model) State() State {
	return m.state
}

// Done reports whether the user has finished with the picker (either selected
// or cancelled).
func (m Model) Done() bool {
	return m.state != StateBrowsing
}

// finish marks the picker as finished. In standalone mode (the Pick* API) it
// ends the program with tea.Quit; when embedded it emits a DoneMsg so the
// parent program can react.
func (m Model) finish(s State) (tea.Model, tea.Cmd) {
	m.state = s
	if m.standalone {
		m.quitting = true
		return m, tea.Quit
	}
	paths := m.SelectedPaths()
	return m, func() tea.Msg { return DoneMsg{State: s, Paths: paths} }
}

// SelectedPath returns the single selected path after the picker closes.
// Returns empty string if nothing was selected.
func (m Model) SelectedPath() string {
	if m.state != StateSelected {
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
	if m.state != StateSelected {
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
	fsys := m.fsys
	showHidden := m.options.ShowHidden
	filters := m.options.Filters
	return func() tea.Msg {
		raw, err := fsys.ReadDir(dir)
		if err != nil {
			return dirReadMsg{dir: dir, err: err}
		}
		entries := buildEntries(fsys, dir, raw, showHidden, filters)
		return dirReadMsg{entries: entries, dir: dir}
	}
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.refreshPreview()
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
		m.refreshPreview()
		return m, nil

	case tea.KeyMsg:
		prev := m.currentPath()
		updated, cmd := m.handleKey(msg)
		nm, ok := updated.(Model)
		if !ok {
			return updated, cmd
		}
		// Refresh the preview only when the selection changed and no directory
		// read is pending (a pending read refreshes via dirReadMsg instead).
		if cmd == nil && nm.currentPath() != prev {
			nm.refreshPreview()
		}
		return nm, cmd
	}

	return m, nil
}

// currentPath returns the path of the highlighted entry, or "" if none.
func (m Model) currentPath() string {
	if m.cursor >= 0 && m.cursor < len(m.entries) {
		return m.entries[m.cursor].Path
	}
	return ""
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
		return m.finish(StateCancelled)

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
		parent := m.resolve(m.fsys.Parent(m.dir))
		if parent != m.dir {
			m.returnTo = m.fsys.Base(m.dir)
			m.dir = parent
			return m, m.readDir()
		}
		m.statusMsg = "At highest level — press q to quit"

	case key.Matches(msg, m.keys.Navigate):
		return m.handleNavigate()

	case key.Matches(msg, m.keys.SelectDir):
		return m.handleSelectDir()

	case key.Matches(msg, m.keys.NewFile) && m.options.Interactive && m.options.Mode != ModeFolder:
		if !m.writable() {
			m.statusMsg = readOnlyMsg
			return m, nil
		}
		m.inputMode = inputNewFile
		m.inputText = ""

	case key.Matches(msg, m.keys.NewFolder) && m.options.Interactive,
		key.Matches(msg, m.keys.NewFile) && m.options.Interactive && m.options.Mode == ModeFolder:
		if !m.writable() {
			m.statusMsg = readOnlyMsg
			return m, nil
		}
		m.inputMode = inputNewFolder
		m.inputText = ""

	case key.Matches(msg, m.keys.Delete) && m.options.Interactive:
		if !m.writable() {
			m.statusMsg = readOnlyMsg
			return m, nil
		}
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
		m.entries = []FileEntry{{
			Name:  ".",
			Path:  m.dir,
			IsDir: true,
		}}
		m.cursor = 0
		return m.finish(StateSelected)
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
		return m.finish(StateSelected)

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
			return m.finish(StateSelected)

		case ModeFolder:
			if entry.IsDir {
				return m.finish(StateSelected)
			}

		case ModeAny:
			if entry.IsDir {
				m.dir = m.resolve(entry.Path)
				return m, m.readDir()
			}
			return m.finish(StateSelected)

		case ModeMultiple:
			if entry.IsDir {
				m.dir = m.resolve(entry.Path)
				return m, m.readDir()
			}
			if len(m.selected) > 0 {
				return m.finish(StateSelected)
			}
			m.selected[entry.Path] = struct{}{}
			return m.finish(StateSelected)
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

		w, ok := m.fsys.(WritableFileSystem)
		if !ok {
			m.inputMode = inputNone
			m.inputText = ""
			m.statusMsg = readOnlyMsg
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
			err = w.CreateDir(m.dir, name)
		} else if m.inputMode == inputNewFile {
			err = w.CreateFile(m.dir, name)
		} else {
			err = w.CreateDir(m.dir, name)
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
		switch msg.Type {
		case tea.KeyRunes:
			m.inputText += string(msg.Runes)
		case tea.KeySpace:
			m.inputText += " "
		}
		return m, nil
	}
}

func (m Model) handleDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.inputMode = inputNone

	if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && (msg.Runes[0] == 'y' || msg.Runes[0] == 'Y') {
		w, ok := m.fsys.(WritableFileSystem)
		if !ok {
			m.statusMsg = readOnlyMsg
			return m, nil
		}
		if m.cursor >= 0 && m.cursor < len(m.entries) {
			entry := m.entries[m.cursor]
			err := w.Remove(entry.Path)
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
// With fuzzy search enabled (the default), matches are ranked best-match-first;
// otherwise a case-insensitive substring match preserves the original order.
func (m *Model) filterEntries() {
	m.matchIdx = nil
	if m.searchTerm == "" {
		m.entries = m.allEntries
		m.cursor = 0
		m.offset = 0
		return
	}
	if m.options.FuzzySearch {
		m.filterFuzzy()
	} else {
		m.filterSubstring()
	}
	m.cursor = 0
	m.offset = 0
}

// filterFuzzy ranks entries by fuzzy match score against the search term and
// records the matched character offsets for highlighting.
func (m *Model) filterFuzzy() {
	names := make([]string, len(m.allEntries))
	for i, e := range m.allEntries {
		names[i] = e.Name
	}
	matches := fuzzy.Find(m.searchTerm, names)
	filtered := make([]FileEntry, 0, len(matches))
	m.matchIdx = make(map[string][]int, len(matches))
	for _, mt := range matches {
		e := m.allEntries[mt.Index]
		filtered = append(filtered, e)
		m.matchIdx[e.Path] = mt.MatchedIndexes
	}
	m.entries = filtered
}

// filterSubstring keeps entries whose name contains the search term,
// preserving the original ordering.
func (m *Model) filterSubstring() {
	term := strings.ToLower(m.searchTerm)
	var filtered []FileEntry
	for _, e := range m.allEntries {
		if strings.Contains(strings.ToLower(e.Name), term) {
			filtered = append(filtered, e)
		}
	}
	m.entries = filtered
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
	// Account for the chrome around the file list so the view never exceeds the
	// terminal height (which would make it scroll/jump on every repaint).
	// Base chrome: title (2 lines) + path (2) + status counter (1), plus one
	// blank bottom row so a full screen does not make the terminal scroll. Some
	// elements are conditional and the help bar wraps with width, so its height
	// is measured rather than assumed.
	chrome := 7
	if m.options.Mode == ModeFolder || m.options.Mode == ModeAny {
		chrome++ // "press s to select" hint
	}
	if m.searching {
		chrome++ // search prompt line
	}
	if m.inputMode != inputNone {
		chrome++ // new file/folder or delete-confirm prompt line
	}
	chrome += len(m.helpLines())

	h := m.height - chrome
	if m.options.Height > 0 && m.options.Height < h {
		h = m.options.Height
	}
	if h < 3 {
		h = 3
	}
	return h
}

// readOnlyMsg is shown when an interactive action is attempted on a
// filesystem that does not support writes (e.g. a custom io/fs.FS).
const readOnlyMsg = "Filesystem is read-only"

// writable reports whether the active filesystem supports create/delete.
func (m Model) writable() bool {
	_, ok := m.fsys.(WritableFileSystem)
	return ok
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
