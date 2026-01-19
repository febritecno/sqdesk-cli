package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ShortcutItem represents a single shortcut
type ShortcutItem struct {
	Key  string
	Desc string
}

// ShortcutCategory represents a category of shortcuts
type ShortcutCategory struct {
	Name  string
	Items []ShortcutItem
}

// Help component for displaying shortcuts
type Help struct {
	visible  bool
	page     int
	maxPages int
	width    int
	height   int
	styles   HelpStyles
}

// HelpStyles holds styling for the help modal
type HelpStyles struct {
	Modal    lipgloss.Style
	Title    lipgloss.Style
	Category lipgloss.Style
	Key      lipgloss.Style
	Desc     lipgloss.Style
	Footer   lipgloss.Style
}

// All shortcuts organized by category
var shortcuts = []ShortcutCategory{
	{
		Name: "🧭 Navigation",
		Items: []ShortcutItem{
			{"F1", "Focus next pane"},
			{"F2", "Focus previous pane"},
			{"↑/↓", "Navigate items"},
			{"←/→", "Switch sidebar sections"},
			{"Enter", "Select/Execute action"},
		},
	},
	{
		Name: "📝 Editor",
		Items: []ShortcutItem{
			{"Ctrl+A", "Select all"},
			{"Ctrl+C", "Copy selection"},
			{"Ctrl+V", "Paste"},
			{"Ctrl+X", "Cut selection"},
			{"Ctrl+Z", "Undo"},
			{"Ctrl+Y", "Redo"},
			{"Ctrl+F", "Find"},
			{"Ctrl+H", "Find & Replace"},
			{"Ctrl+L", "Go to line"},
		},
	},
	{
		Name: "🔍 Query",
		Items: []ShortcutItem{
			{"F5", "Run query"},
			{"Ctrl+E", "Execute query"},
			{"F3", "Toggle Keywords panel"},
			{"Tab", "Accept suggestion"},
		},
	},
	{
		Name: "🤖 AI Features",
		Items: []ShortcutItem{
			{"Ctrl+G", "AI Generate (Text-to-SQL)"},
			{"Ctrl+K", "AI Refactor query"},
		},
	},
	{
		Name: "📊 Results",
		Items: []ShortcutItem{
			{"c", "Copy selected row"},
			{"C", "Copy all data"},
			{"v", "Toggle chart view"},
			{"1/2/3", "Switch chart type"},
		},
	},
	{
		Name: "⚙️ General",
		Items: []ShortcutItem{
			{"F4", "Show/Hide this help"},
			{"Ctrl+Q", "Quit application"},
			{"Esc", "Close modal/Cancel"},
		},
	},
}

// NewHelp creates a new Help component
func NewHelp(styles HelpStyles) Help {
	return Help{
		visible:  false,
		page:     0,
		maxPages: len(shortcuts),
		styles:   styles,
	}
}

// Show displays the help modal
func (h *Help) Show() {
	h.visible = true
	h.page = 0
}

// Hide hides the help modal
func (h *Help) Hide() {
	h.visible = false
}

// Toggle toggles the help modal visibility
func (h *Help) Toggle() {
	h.visible = !h.visible
	if h.visible {
		h.page = 0
	}
}

// IsVisible returns if help is visible
func (h Help) IsVisible() bool {
	return h.visible
}

// NextPage moves to next page
func (h *Help) NextPage() {
	if h.page < h.maxPages-1 {
		h.page++
	} else {
		h.page = 0 // Wrap around
	}
}

// PrevPage moves to previous page
func (h *Help) PrevPage() {
	if h.page > 0 {
		h.page--
	} else {
		h.page = h.maxPages - 1 // Wrap around
	}
}

// SetSize sets the help modal size
func (h *Help) SetSize(width, height int) {
	h.width = width
	h.height = height
}

// Update handles input
func (h Help) Update(msg tea.Msg) (Help, tea.Cmd) {
	if !h.visible {
		return h, nil
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "left", "h":
			h.PrevPage()
		case "right", "l":
			h.NextPage()
		case "esc", "f4", "q":
			h.Hide()
		}
	}

	return h, nil
}

// View renders the help modal
func (h Help) View() string {
	if !h.visible {
		return ""
	}

	var content strings.Builder

	// Title
	title := h.styles.Title.Render("⌨️  Keyboard Shortcuts")
	content.WriteString(title)
	content.WriteString("\n\n")

	// Current category
	if h.page < len(shortcuts) {
		cat := shortcuts[h.page]
		
		// Category name
		catName := h.styles.Category.Render(cat.Name)
		content.WriteString(catName)
		content.WriteString("\n\n")

		// Shortcuts
		for _, item := range cat.Items {
			key := h.styles.Key.Width(15).Render(item.Key)
			desc := h.styles.Desc.Render(item.Desc)
			content.WriteString(fmt.Sprintf("  %s  %s\n", key, desc))
		}
	}

	// Footer with pagination
	content.WriteString("\n")
	pagination := fmt.Sprintf("← → Page %d/%d", h.page+1, h.maxPages)
	footer := h.styles.Footer.Render(pagination + "  •  Press Esc to close")
	content.WriteString(footer)

	return h.styles.Modal.Render(content.String())
}
