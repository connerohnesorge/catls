// Package interactive provides a TUI file selector using bubbletea.
package interactive

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FileItem represents a file in the selector.
type FileItem struct {
	Path     string
	RelPath  string
	IsBinary bool
	Selected bool
}

// KeyMap defines the keybindings for the selector.
type KeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Toggle      key.Binding
	SelectAll   key.Binding
	DeselectAll key.Binding
	Confirm     key.Binding
	Quit        key.Binding
}

// DefaultKeyMap returns the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Toggle: key.NewBinding(
			key.WithKeys(" ", "x"),
			key.WithHelp("space/x", "toggle"),
		),
		SelectAll: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "select all"),
		),
		DeselectAll: key.NewBinding(
			key.WithKeys("A"),
			key.WithHelp("A", "deselect all"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c", "esc"),
			key.WithHelp("q/esc", "quit"),
		),
	}
}

// Model is the bubbletea model for the file selector.
type Model struct {
	files     []FileItem
	cursor    int
	keys      KeyMap
	viewport  viewport.Model
	ready     bool
	width     int
	height    int
	quitting  bool
	confirmed bool
}

// NewModel creates a new file selector model.
func NewModel(files []FileItem) Model {
	return Model{
		files: files,
		keys:  DefaultKeyMap(),
	}
}

// SelectedFiles returns the list of selected files.
func (m *Model) SelectedFiles() []FileItem {
	var selected []FileItem
	for _, f := range m.files {
		if f.Selected {
			selected = append(selected, f)
		}
	}

	return selected
}

// Init implements tea.Model.
func (*Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model. It dispatches input to typed handlers and
// forwards any residual message to the embedded viewport.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.KeyMsg:
		if cmd, done := m.handleKey(typed); done {
			return m, cmd
		}
	case tea.WindowSizeMsg:
		m.resize(typed.Width, typed.Height)

		return m, nil
	}

	if !m.ready {
		return m, nil
	}

	m.viewport.SetContent(m.renderContent())

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)

	return m, cmd
}

// handleKey routes a key event. The returned bool is true when the caller
// should immediately return with the supplied command.
func (m *Model) handleKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		m.quitting = true

		return tea.Quit, true
	case key.Matches(msg, m.keys.Confirm):
		m.confirmed = true

		return tea.Quit, true
	case key.Matches(msg, m.keys.Up):
		m.moveCursor(-1)
	case key.Matches(msg, m.keys.Down):
		m.moveCursor(1)
	case key.Matches(msg, m.keys.Toggle):
		m.toggleCurrent()
	case key.Matches(msg, m.keys.SelectAll):
		m.setAll(true)
	case key.Matches(msg, m.keys.DeselectAll):
		m.setAll(false)
	}

	return nil, false
}

// moveCursor shifts the cursor by delta, clamped to the list bounds.
func (m *Model) moveCursor(delta int) {
	next := m.cursor + delta
	if next < 0 || next >= len(m.files) {
		return
	}
	m.cursor = next
	m.ensureCursorVisible()
}

// toggleCurrent flips the selection state of the row under the cursor.
func (m *Model) toggleCurrent() {
	if len(m.files) == 0 {
		return
	}
	m.files[m.cursor].Selected = !m.files[m.cursor].Selected
}

// setAll sets every file's selected flag to the same value.
func (m *Model) setAll(selected bool) {
	for i := range m.files {
		m.files[i].Selected = selected
	}
}

// resize initializes or updates the internal viewport for the given size.
func (m *Model) resize(w, h int) {
	m.width = w
	m.height = h
	m.viewport = viewport.New(w, h-4)
	m.viewport.SetContent(m.renderContent())
	m.ready = true
}

// ensureCursorVisible scrolls the viewport so the cursor row stays on screen.
func (m *Model) ensureCursorVisible() {
	if !m.ready {
		return
	}
	itemHeight := 1
	cursorY := m.cursor * itemHeight
	viewportBottom := m.viewport.YOffset + m.viewport.Height

	if cursorY < m.viewport.YOffset {
		m.viewport.YOffset = cursorY
	} else if cursorY >= viewportBottom {
		m.viewport.YOffset = cursorY - m.viewport.Height + itemHeight
	}
}

// Styling used to render the selector. Colors are standard 16-color indices so
// the UI adapts to the user's terminal theme.
var (
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	binaryStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	normalStyle   = lipgloss.NewStyle()
)

// View implements tea.Model.
func (m *Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	if m.quitting {
		return "Cancelled.\n"
	}

	if m.confirmed {
		selected := m.SelectedFiles()

		return fmt.Sprintf("Selected %d file(s).\n", len(selected))
	}

	selectedCount := len(m.SelectedFiles())
	header := headerStyle.Render(fmt.Sprintf("Select files (selected: %d/%d)", selectedCount, len(m.files)))
	content := m.viewport.View()
	footer := fmt.Sprintf(
		"%s %s %s %s %s %s",
		m.renderKeyHelp(m.keys.Up),
		m.renderKeyHelp(m.keys.Down),
		m.renderKeyHelp(m.keys.Toggle),
		m.renderKeyHelp(m.keys.SelectAll),
		m.renderKeyHelp(m.keys.DeselectAll),
		m.renderKeyHelp(m.keys.Confirm),
	)

	return fmt.Sprintf("%s\n%s\n%s", header, content, dimStyle.Render(footer))
}

// renderKeyHelp formats a single key binding for the footer help strip.
func (*Model) renderKeyHelp(k key.Binding) string {
	return fmt.Sprintf("[%s %s]", k.Keys()[0], k.Help().Desc)
}

// renderContent produces the scrollable body: one row per file with cursor,
// checkbox, and path (dimmed for binaries).
func (m *Model) renderContent() string {
	var b strings.Builder

	for i, file := range m.files {
		cursor := " "
		if i == m.cursor {
			cursor = cursorStyle.Render(">")
		}

		checkbox := "[ ]"
		if file.Selected {
			checkbox = selectedStyle.Render("[x]")
		}

		path := file.RelPath
		style := normalStyle
		if file.IsBinary {
			style = binaryStyle
			path = file.RelPath + dimStyle.Render(" (binary)")
		}

		b.WriteString(fmt.Sprintf("%s %s %s\n", cursor, checkbox, style.Render(path)))
	}

	return b.String()
}

// SelectFiles launches the interactive file selector and returns the selected files.
// Returns nil if the user cancels or no files are selected.
func SelectFiles(files []FileItem) ([]FileItem, error) {
	if len(files) == 0 {
		return nil, nil
	}

	for i := range files {
		files[i].Selected = true
	}

	m := NewModel(files)
	p := tea.NewProgram(&m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return nil, fmt.Errorf("failed to run selector: %w", err)
	}

	if m.quitting || !m.confirmed {
		return nil, nil
	}

	selected := m.SelectedFiles()
	if len(selected) == 0 {
		return nil, nil
	}

	return selected, nil
}
