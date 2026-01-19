package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the entire application
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Handle setup wizard
	if m.state == StateSetup {
		return m.wizard.View()
	}

	var content strings.Builder

	// Header
	content.WriteString(m.renderHeader())
	content.WriteString("\n")

	// Main content
	content.WriteString(m.renderMainContent())

	// Footer
	content.WriteString("\n")
	content.WriteString(m.renderFooter())

	baseView := content.String()

	// Overlay modals using lipgloss.Place for proper centering
	if m.state == StateAIPrompt && m.aiPrompt.IsVisible() {
		modalContent := m.aiPrompt.View()
		baseView = lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			modalContent,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#333333")),
		)
	}

	if m.state == StateSettings && m.settings.IsVisible() {
		modalContent := m.settings.View()
		baseView = lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			modalContent,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#333333")),
		)
	}
	
	if m.state == StateConnModal && m.connModal.IsVisible() {
		modalContent := m.connModal.View()
		baseView = lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			modalContent,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#333333")),
		)
	}
	
	// Render Help modal if visible
	if m.help.IsVisible() {
		modalContent := m.help.View()
		baseView = lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			modalContent,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#333333")),
		)
	}

	return baseView
}

// renderHeader renders the top header bar
func (m *Model) renderHeader() string {
	// Left: App title
	title := m.styles.Header.Render(" SQDesk ")

	// Center: Connection status
	var connStatus string
	if m.isConnected {
		connStatus = m.styles.SuccessText.Render("● Connected")
	} else {
		connStatus = m.styles.ErrorText.Render("○ Disconnected")
	}

	// Right: AI info
	aiInfo := m.GetAIInfo()
	ai := m.styles.StatusItem.Render(aiInfo)

	// Calculate spacing
	leftWidth := lipgloss.Width(title)
	centerWidth := lipgloss.Width(connStatus)
	rightWidth := lipgloss.Width(ai)
	
	totalWidth := m.width
	spacing := totalWidth - leftWidth - centerWidth - rightWidth
	if spacing < 0 {
		spacing = 2
	}
	leftSpacing := spacing / 2
	rightSpacing := spacing - leftSpacing

	header := title + strings.Repeat(" ", leftSpacing) + connStatus + strings.Repeat(" ", rightSpacing) + ai

	return m.styles.Header.Width(m.width).Render(header)
}

// renderMainContent renders the main workspace
func (m *Model) renderMainContent() string {
	// Calculate dimensions
	sidebarWidth := 20
	if m.width < 80 {
		sidebarWidth = 15
	}
	
	// Right panel for completion (only when visible)
	rightPanelWidth := 0
	if m.completion.IsVisible() {
		rightPanelWidth = 35
	}

	mainWidth := m.width - sidebarWidth - rightPanelWidth - 1
	contentHeight := m.height - 4 // header + footer + margins

	editorHeight := contentHeight * 45 / 100
	resultsHeight := contentHeight - editorHeight - 1

	// Update component sizes
	m.sidebar.SetSize(sidebarWidth, contentHeight)
	m.editor.SetSize(mainWidth, editorHeight)
	m.results.SetSize(mainWidth, resultsHeight)

	// Render sidebar
	sidebar := m.renderSidebar(sidebarWidth, contentHeight)

	// Render editor
	editor := m.renderEditor(mainWidth, editorHeight)

	// Render results
	results := m.renderResults(mainWidth, resultsHeight)

	// Combine editor and results vertically
	centerPane := lipgloss.JoinVertical(lipgloss.Left, editor, results)

	// If completion is visible, add right panel
	if m.completion.IsVisible() {
		m.completion.SetWidth(rightPanelWidth - 2)
		completionPane := m.renderCompletionPane(rightPanelWidth, contentHeight)
		return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, centerPane, completionPane)
	}

	// Combine sidebar and center pane horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, centerPane)
}

// renderSidebar renders the sidebar with connections and tables
func (m *Model) renderSidebar(width, height int) string {
	m.sidebar.SetSize(width, height)
	return m.sidebar.View()
}

// renderEditor renders the SQL editor
func (m *Model) renderEditor(width, height int) string {
	m.editor.SetSize(width, height)
	return m.editor.View()
}

// renderCompletionPane renders the right-side completion panel
func (m *Model) renderCompletionPane(width, height int) string {
	title := m.styles.PanelTitle.Render("KEYWORDS")
	popup := m.completion.View()
	
	content := title + "\n" + popup
	
	style := m.styles.Panel
	return style.
		Width(width).
		Height(height).
		Render(content)
}

// renderResults renders the query results
func (m *Model) renderResults(width, height int) string {
	m.results.SetSize(width, height)
	return m.results.View()
}

// renderFooter renders the bottom status bar
func (m *Model) renderFooter() string {
	// Help text
	helpItems := []struct {
		key  string
		desc string
	}{
		{"F5", "Run"},
		{"Ctrl+G", "AI"},
		{"Ctrl+K", "Refactor"},
		{"F1/F2", "Switch Pane"},
		{"F3", "Keywords Panel"},
		{"F4", "Help"},
		{"Ctrl+Q", "Quit"},
	}

	var help strings.Builder
	for i, item := range helpItems {
		help.WriteString(m.styles.HelpKey.Render(item.key))
		help.WriteString(" ")
		help.WriteString(m.styles.HelpDesc.Render(item.desc))
		if i < len(helpItems)-1 {
			help.WriteString("  ")
		}
	}

	// Status message
	var status string
	if m.statusMessage != "" {
		if m.isError {
			status = m.styles.ErrorText.Render(m.statusMessage)
		} else {
			status = m.styles.SuccessText.Render(m.statusMessage)
		}
	}

	footer := help.String()
	if status != "" {
		spacing := m.width - lipgloss.Width(help.String()) - lipgloss.Width(status) - 2
		if spacing > 0 {
			footer = footer + strings.Repeat(" ", spacing) + status
		}
	}

	return m.styles.Footer.Width(m.width).Render(footer)
}

// highlightSQL applies basic SQL syntax highlighting
func (m *Model) highlightSQL(line string) string {
	keywords := []string{
		"SELECT", "FROM", "WHERE", "AND", "OR", "NOT", "IN", "LIKE",
		"ORDER", "BY", "ASC", "DESC", "LIMIT", "OFFSET", "GROUP",
		"HAVING", "JOIN", "LEFT", "RIGHT", "INNER", "OUTER", "ON",
		"INSERT", "INTO", "VALUES", "UPDATE", "SET", "DELETE",
		"CREATE", "TABLE", "DROP", "ALTER", "INDEX", "VIEW",
		"AS", "DISTINCT", "COUNT", "SUM", "AVG", "MIN", "MAX",
		"NULL", "IS", "BETWEEN", "EXISTS", "CASE", "WHEN", "THEN",
		"ELSE", "END", "UNION", "ALL", "TRUE", "FALSE",
	}

	result := line
	for _, kw := range keywords {
		// Case insensitive replacement
		upper := strings.ToUpper(result)
		lower := strings.ToLower(kw)
		
		// Find and highlight keywords
		idx := 0
		for {
			pos := strings.Index(strings.ToUpper(result[idx:]), kw)
			if pos == -1 {
				break
			}
			
			actualPos := idx + pos
			// Check word boundaries
			before := actualPos == 0 || !isAlphaNum(result[actualPos-1])
			after := actualPos+len(kw) >= len(result) || !isAlphaNum(result[actualPos+len(kw)])
			
			if before && after {
				// Replace with highlighted version
				highlighted := m.styles.Keyword.Render(result[actualPos : actualPos+len(kw)])
				result = result[:actualPos] + highlighted + result[actualPos+len(kw):]
				idx = actualPos + len(highlighted)
			} else {
				idx = actualPos + len(kw)
			}
			
			_ = upper
			_ = lower
		}
	}

	return result
}

// Helper functions

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func isAlphaNum(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}
