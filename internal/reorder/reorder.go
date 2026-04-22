// Package reorder provides a TUI for manually reordering a list of files
// before output. It mirrors the style of the interactive selector.
package reorder

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/connerohnesorge/catls/internal/scanner"
)

// KeyMap defines the keybindings for the reorder TUI.
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Grab     key.Binding
	MoveUp   key.Binding
	MoveDown key.Binding
	Confirm  key.Binding
	Quit     key.Binding
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
		Grab: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "grab/drop"),
		),
		MoveUp: key.NewBinding(
			key.WithKeys("shift+up", "K"),
			key.WithHelp("shift+↑/K", "move up"),
		),
		MoveDown: key.NewBinding(
			key.WithKeys("shift+down", "J"),
			key.WithHelp("shift+↓/J", "move down"),
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

// Model is the bubbletea model for the reorder TUI.
type Model struct {
	files     []scanner.FileInfo
	cursor    int
	grabbed   bool
	keys      KeyMap
	viewport  viewport.Model
	ready     bool
	width     int
	height    int
	quitting  bool
	confirmed bool
}

// NewModel creates a new reorder model seeded with the given files.
func NewModel(files []scanner.FileInfo) Model {
	ordered := make([]scanner.FileInfo, len(files))
	copy(ordered, files)

	return Model{
		files: ordered,
		keys:  DefaultKeyMap(),
	}
}

// Files returns the current ordering of the files.
func (m *Model) Files() []scanner.FileInfo {
	out := make([]scanner.FileInfo, len(m.files))
	copy(out, m.files)

	return out
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
	case key.Matches(msg, m.keys.Grab):
		if len(m.files) > 0 {
			m.grabbed = !m.grabbed
		}
	case key.Matches(msg, m.keys.MoveUp):
		m.moveItemUp()
	case key.Matches(msg, m.keys.MoveDown):
		m.moveItemDown()
	case key.Matches(msg, m.keys.Up):
		m.handleUp()
	case key.Matches(msg, m.keys.Down):
		m.handleDown()
	}

	return nil, false
}

// resize initializes or updates the internal viewport for the given size.
func (m *Model) resize(w, h int) {
	m.width = w
	m.height = h
	m.viewport = viewport.New(w, h-4)
	m.viewport.SetContent(m.renderContent())
	m.ready = true
}

// handleUp handles a plain up-arrow: moves the grabbed item or the cursor.
func (m *Model) handleUp() {
	if m.grabbed {
		m.moveItemUp()

		return
	}
	if m.cursor > 0 {
		m.cursor--
		m.ensureCursorVisible()
	}
}

// handleDown is the down-arrow counterpart to handleUp.
func (m *Model) handleDown() {
	if m.grabbed {
		m.moveItemDown()

		return
	}
	if m.cursor < len(m.files)-1 {
		m.cursor++
		m.ensureCursorVisible()
	}
}

// moveItemUp swaps the item at the cursor with the one above it.
func (m *Model) moveItemUp() {
	if m.cursor <= 0 || len(m.files) == 0 {
		return
	}
	m.files[m.cursor-1], m.files[m.cursor] = m.files[m.cursor], m.files[m.cursor-1]
	m.cursor--
	m.ensureCursorVisible()
}

// moveItemDown swaps the item at the cursor with the one below it.
func (m *Model) moveItemDown() {
	if m.cursor >= len(m.files)-1 || len(m.files) == 0 {
		return
	}
	m.files[m.cursor+1], m.files[m.cursor] = m.files[m.cursor], m.files[m.cursor+1]
	m.cursor++
	m.ensureCursorVisible()
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

// Styling used to render the reorder TUI. The grabbed style is deliberately
// distinct from the cursor style so the current mode is unambiguous.
var (
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	cursorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	grabbedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true)
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	binaryStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	normalStyle  = lipgloss.NewStyle()
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
		return fmt.Sprintf("Reordered %d file(s).\n", len(m.files))
	}

	mode := "browse"
	if m.grabbed {
		mode = grabbedStyle.Render("GRABBED")
	}
	header := headerStyle.Render(fmt.Sprintf("Reorder files (%d) — mode: %s", len(m.files), mode))
	content := m.viewport.View()
	footer := fmt.Sprintf(
		"%s %s %s %s %s %s",
		m.renderKeyHelp(m.keys.Up),
		m.renderKeyHelp(m.keys.Down),
		m.renderKeyHelp(m.keys.Grab),
		m.renderKeyHelp(m.keys.MoveUp),
		m.renderKeyHelp(m.keys.MoveDown),
		m.renderKeyHelp(m.keys.Confirm),
	)

	return fmt.Sprintf("%s\n%s\n%s", header, content, dimStyle.Render(footer))
}

// renderKeyHelp formats a single key binding for the footer help strip.
func (*Model) renderKeyHelp(k key.Binding) string {
	return fmt.Sprintf("[%s %s]", k.Keys()[0], k.Help().Desc)
}

// renderContent produces the scrollable body: one row per file, numbered and
// annotated with cursor/grab state and a binary marker.
func (m *Model) renderContent() string {
	var b strings.Builder

	for i, file := range m.files {
		cursor := " "
		if i == m.cursor {
			if m.grabbed {
				cursor = grabbedStyle.Render("#")
			} else {
				cursor = cursorStyle.Render(">")
			}
		}

		path := file.RelPath
		style := normalStyle
		if file.IsBinary {
			style = binaryStyle
			path = file.RelPath + dimStyle.Render(" (binary)")
		}
		if i == m.cursor && m.grabbed {
			style = grabbedStyle
		}

		b.WriteString(fmt.Sprintf("%s %3d  %s\n", cursor, i+1, style.Render(path)))
	}

	return b.String()
}

// Reorder launches the reorder TUI and returns the new ordering.
// Returns nil if the user cancels.
func Reorder(files []scanner.FileInfo) ([]scanner.FileInfo, error) {
	if len(files) == 0 {
		return files, nil
	}

	m := NewModel(files)
	p := tea.NewProgram(&m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return nil, fmt.Errorf("failed to run reorder TUI: %w", err)
	}

	if m.quitting || !m.confirmed {
		return nil, nil
	}

	return m.Files(), nil
}
