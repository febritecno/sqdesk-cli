package components

import (
	"github.com/charmbracelet/lipgloss"
)

// ConnectionAction represents the action selected in the modal
type ConnectionAction int

const (
	ActionNone ConnectionAction = iota
	ActionConnect
	ActionEdit
	ActionDelete
	ActionCancel
)

// ConnectionModal component for connection actions pop-up
type ConnectionModal struct {
	visible        bool
	width          int
	height         int
	connName       string
	connIndex      int
	selectedAction int
	actions        []string
	status         string
	isError        bool
	styles         ConnectionModalStyles
}

// ConnectionModalStyles holds styling for the modal
type ConnectionModalStyles struct {
	Modal    lipgloss.Style
	Title    lipgloss.Style
	Item     lipgloss.Style
	Selected lipgloss.Style
	Hint     lipgloss.Style
	Success  lipgloss.Style
	Error    lipgloss.Style
}

// NewConnectionModal creates a new connection modal
func NewConnectionModal(styles ConnectionModalStyles) ConnectionModal {
	return ConnectionModal{
		visible:        false,
		selectedAction: 0,
		actions:        []string{"🔗 Connect", "✏️  Edit", "🗑️  Delete", "✕  Cancel"},
		styles:         styles,
	}
}

// Show shows the modal for a connection
func (m *ConnectionModal) Show(connName string, connIndex int) {
	m.visible = true
	m.connName = connName
	m.connIndex = connIndex
	m.selectedAction = 0
	m.status = ""
	m.isError = false
}

// SetStatus sets the status message
func (m *ConnectionModal) SetStatus(msg string, isError bool) {
	m.status = msg
	m.isError = isError
}

// Hide hides the modal
func (m *ConnectionModal) Hide() {
	m.visible = false
}

// IsVisible returns if modal is visible
func (m ConnectionModal) IsVisible() bool {
	return m.visible
}

// GetConnectionIndex returns the connection index
func (m ConnectionModal) GetConnectionIndex() int {
	return m.connIndex
}

// SetSize sets the modal dimensions
func (m *ConnectionModal) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// MoveUp moves selection up
func (m *ConnectionModal) MoveUp() {
	if m.selectedAction > 0 {
		m.selectedAction--
	}
}

// MoveDown moves selection down
func (m *ConnectionModal) MoveDown() {
	if m.selectedAction < len(m.actions)-1 {
		m.selectedAction++
	}
}

// GetSelectedAction returns the selected action
func (m ConnectionModal) GetSelectedAction() ConnectionAction {
	switch m.selectedAction {
	case 0:
		return ActionConnect
	case 1:
		return ActionEdit
	case 2:
		return ActionDelete
	case 3:
		return ActionCancel
	default:
		return ActionNone
	}
}

// View renders the modal
func (m ConnectionModal) View() string {
	if !m.visible {
		return ""
	}

	// Title
	title := m.styles.Title.Render("📁 " + m.connName)

	// Build action list
	content := title + "\n\n"

	for i, action := range m.actions {
		style := m.styles.Item
		if i == m.selectedAction {
			style = m.styles.Selected
		}
		content += style.Render(action) + "\n"
	}
	
	// Status message
	if m.status != "" {
		statusStyle := m.styles.Success
		if m.isError {
			statusStyle = m.styles.Error
		}
		content += "\n" + statusStyle.Render(m.status)
	}

	content += "\n" + m.styles.Hint.Render("↑↓: navigate • Enter: select • Esc: cancel")

	// Ensure minimum width
	width := m.width
	if width < 40 {
		width = 40
	}

	return m.styles.Modal.
		Width(width).
		Padding(1, 2).
		Render(content)
}
