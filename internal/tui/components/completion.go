package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/febritecno/sqdesk-cli/internal/completion"
)

// CompletionPopup shows completion suggestions
type CompletionPopup struct {
	items    []completion.CompletionItem
	selected int
	visible  bool
	x, y     int // Position
	width    int
	maxItems int
	styles   CompletionStyles
}

// CompletionStyles holds styling for the popup
type CompletionStyles struct {
	Border      lipgloss.Style
	Item        lipgloss.Style
	Selected    lipgloss.Style
	Kind        lipgloss.Style
	Detail      lipgloss.Style
}

// NewCompletionPopup creates a new completion popup
func NewCompletionPopup(styles CompletionStyles) CompletionPopup {
	return CompletionPopup{
		items:    make([]completion.CompletionItem, 0),
		selected: 0,
		visible:  false,
		maxItems: 10,
		width:    40,
		styles:   styles,
	}
}

// Show displays the popup with items
func (c *CompletionPopup) Show(items []completion.CompletionItem, x, y int) {
	c.items = items
	c.selected = 0
	c.x = x
	c.y = y
	c.visible = len(items) > 0
}

// Hide hides the popup
func (c *CompletionPopup) Hide() {
	c.visible = false
	c.items = nil
	c.selected = 0
}

// IsVisible returns if popup is visible
func (c CompletionPopup) IsVisible() bool {
	return c.visible
}

// MoveUp moves selection up
func (c *CompletionPopup) MoveUp() {
	if c.selected > 0 {
		c.selected--
	}
}

// MoveDown moves selection down
func (c *CompletionPopup) MoveDown() {
	if c.selected < len(c.items)-1 {
		c.selected++
	}
}

// GetSelected returns the selected item
func (c CompletionPopup) GetSelected() *completion.CompletionItem {
	if c.selected >= 0 && c.selected < len(c.items) {
		return &c.items[c.selected]
	}
	return nil
}

// Update handles input
func (c CompletionPopup) Update(msg tea.Msg) (CompletionPopup, tea.Cmd) {
	if !c.visible {
		return c, nil
	}
	
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "up", "ctrl+p":
			c.MoveUp()
		case "down", "ctrl+n":
			c.MoveDown()
		case "esc":
			c.Hide()
		}
	}
	
	return c, nil
}

// View renders the popup
func (c CompletionPopup) View() string {
	if !c.visible || len(c.items) == 0 {
		return ""
	}
	
	var content strings.Builder
	
	// Determine visible range
	start := 0
	end := len(c.items)
	if end > c.maxItems {
		// Scroll to keep selected visible
		if c.selected >= c.maxItems {
			start = c.selected - c.maxItems + 1
		}
		end = start + c.maxItems
		if end > len(c.items) {
			end = len(c.items)
		}
	}
	
	for i := start; i < end; i++ {
		item := c.items[i]
		
		// Build line
		icon := item.Kind.Icon()
		label := truncateLabel(item.Label, c.width-10)
		line := icon + " " + label
		
		// Add detail if space allows
		if len(line) < c.width-10 && item.Detail != "" {
			detail := truncateLabel(item.Detail, c.width-len(line)-5)
			line += " " + c.styles.Detail.Render(detail)
		}
		
		// Apply style
		style := c.styles.Item
		if i == c.selected {
			style = c.styles.Selected
		}
		
		content.WriteString(style.Width(c.width).Render(line))
		if i < end-1 {
			content.WriteString("\n")
		}
	}
	
	// Scroll indicator
	if len(c.items) > c.maxItems {
		scrollInfo := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Render(" ↑↓ scroll")
		content.WriteString("\n" + scrollInfo)
	}
	
	return c.styles.Border.Render(content.String())
}

// SetWidth sets the popup width
func (c *CompletionPopup) SetWidth(w int) {
	c.width = w
}

// SetPosition sets the popup position
func (c *CompletionPopup) SetPosition(x, y int) {
	c.x = x
	c.y = y
}

// GetPosition returns the popup position
func (c CompletionPopup) GetPosition() (int, int) {
	return c.x, c.y
}

// truncateLabel shortens text for display
func truncateLabel(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
