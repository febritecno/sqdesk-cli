package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/febritecno/sqdesk-cli/internal/ai"
	"github.com/febritecno/sqdesk-cli/internal/config"
	"github.com/febritecno/sqdesk-cli/internal/tui/components"
)

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles all input and state changes
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		return m, nil

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case tea.KeyMsg:
		// Handle global keys first
		cmd := m.handleGlobalKeys(msg)
		if cmd != nil {
			return m, cmd
		}

		// Handle state-specific keys
		switch m.state {
		case StateSetup:
			return m.updateSetup(msg)
		case StateAIPrompt:
			return m.updateAIPrompt(msg)
		case StateSettings:
			return m.updateSettings(msg)
		case StateConnModal:
			return m.updateConnModal(msg)
		case StateNormal:
			return m.updateNormal(msg)
		}
	}

	return m, tea.Batch(cmds...)
}

// handleMouse handles mouse events
func (m *Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Skip if in modal state
	if m.state != StateNormal {
		return m, nil
	}

	// Calculate pane boundaries
	sidebarWidth := 20
	if m.width < 80 {
		sidebarWidth = 15
	}
	headerHeight := 2
	footerHeight := 2
	editorHeight := (m.height - headerHeight - footerHeight) * 40 / 100

	// Determine which pane was clicked
	x, y := msg.X, msg.Y

	switch msg.Type {
	case tea.MouseLeft:
		// Check if click is in sidebar
		if x < sidebarWidth && y >= headerHeight && y < m.height-footerHeight {
			m.focusedPane = PaneSidebar
			m.sidebar.SetFocused(true)
			m.editor.SetFocused(false)
			m.results.SetFocused(false)
			
			// Handle sidebar click - determine section and item
			m.handleSidebarClick(y - headerHeight)
			return m, nil
		}

		// Check if click is in editor
		if x >= sidebarWidth && y >= headerHeight && y < headerHeight+editorHeight {
			m.focusedPane = PaneEditor
			m.sidebar.SetFocused(false)
			m.editor.SetFocused(true)
			m.results.SetFocused(false)
			return m, nil
		}

		// Check if click is in results
		if x >= sidebarWidth && y >= headerHeight+editorHeight && y < m.height-footerHeight {
			m.focusedPane = PaneResults
			m.sidebar.SetFocused(false)
			m.editor.SetFocused(false)
			m.results.SetFocused(true)
			return m, nil
		}

	case tea.MouseWheelUp:
		// Scroll up in focused pane
		if m.focusedPane == PaneSidebar {
			var cmd tea.Cmd
			m.sidebar, cmd = m.sidebar.Update(msg)
			return m, cmd
		} else if m.focusedPane == PaneResults {
			m.results.PrevPage()
		}
		return m, nil

	case tea.MouseWheelDown:
		// Scroll down in focused pane
		if m.focusedPane == PaneSidebar {
			var cmd tea.Cmd
			m.sidebar, cmd = m.sidebar.Update(msg)
			return m, cmd
		} else if m.focusedPane == PaneResults {
			m.results.NextPage()
		}
		return m, nil
	}

	return m, nil
}

// handleSidebarClick handles click inside sidebar
func (m *Model) handleSidebarClick(relY int) {
	// Mouse click selection disabled for now due to list component integration
}

// FocusEditor focuses the editor pane
func (m *Model) FocusEditor() {
	m.focusedPane = PaneEditor
	m.sidebar.SetFocused(false)
	m.editor.SetFocused(true)
	m.results.SetFocused(false)
}

// handleGlobalKeys handles keys that work in any state
func (m *Model) handleGlobalKeys(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+c", "ctrl+q":
		return tea.Quit
	}
	return nil
}

// updateSetup handles setup wizard state
func (m *Model) updateSetup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		cmd := m.wizard.Update(msg)
		if m.wizard.IsComplete() {
			// Apply wizard config
			m.config = m.wizard.GetConfig()
			m.config.Save()
			m.state = StateNormal
			
			// Try to connect
			if err := m.Connect(); err != nil {
				m.statusMessage = "Connection failed: " + err.Error()
				m.isError = true
			}
			
			// Reinitialize with new theme
			m.UpdateTheme(m.config.Theme)
		}
		return m, cmd
	default:
		cmd := m.wizard.Update(msg)
		return m, cmd
	}
}

// updateAIPrompt handles AI prompt modal state
func (m *Model) updateAIPrompt(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.aiPrompt.Hide()
		m.aiPrompt.ClearContext()
		m.state = StateNormal
		return m, nil
	case "enter":
		prompt := m.aiPrompt.GetValue()
		if prompt != "" {
			if m.aiPrompt.GetMode() == components.AIPromptModeNL2SQL {
				m.GenerateSQL(prompt)
			} else {
				m.RefactorSQL(prompt)
			}
		}
		m.aiPrompt.Hide()
		m.aiPrompt.ClearContext()
		m.state = StateNormal
		return m, nil
	default:
		var cmd tea.Cmd
		m.aiPrompt, cmd = m.aiPrompt.Update(msg)
		return m, cmd
	}
}

// updateConnModal handles connection modal state
func (m *Model) updateConnModal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.connModal.Hide()
		m.state = StateNormal
		return m, nil
	case "up", "k":
		m.connModal.MoveUp()
		return m, nil
	case "down", "j":
		m.connModal.MoveDown()
		return m, nil
	case "enter":
		action := m.connModal.GetSelectedAction()
		connIdx := m.connModal.GetConnectionIndex()
		m.connModal.Hide()
		m.state = StateNormal
		
		switch action {
		case components.ActionConnect:
			// Connect to selected connection
			if connIdx >= 0 && connIdx < len(m.config.Connections) {
				// Test connection first
				conn := m.config.Connections[connIdx]
				m.connModal.SetStatus("Connecting...", false)
				
				// We need to run this async or just block briefly since it's a TUI
				// For now, we'll block briefly as we don't have async msg handling set up for this yet
				if err := m.TestConnection(&conn); err != nil {
					m.connModal.SetStatus("Connection failed: "+err.Error(), true)
					return m, nil
				}
				
				// If test passes, proceed to connect
				m.config.ActiveConnIndex = connIdx
				m.sidebar.SetActiveConnection(connIdx)
				if m.connector != nil {
					m.connector.Close()
				}
				if err := m.Connect(); err != nil {
					m.connModal.SetStatus("Connect error: "+err.Error(), true)
					return m, nil
				} else {
					m.statusMessage = "Connected to " + m.config.Connections[connIdx].Name
					m.isError = false
					m.connModal.Hide()
					m.state = StateNormal
				}
			}
		case components.ActionEdit:
			// Load connection for editing
			if connIdx >= 0 && connIdx < len(m.config.Connections) {
				conn := m.config.Connections[connIdx]
				m.settings.LoadConnection(conn.Name, conn.Driver, conn.Host, conn.Port, conn.User, conn.Password, conn.Database, connIdx)
				m.settings.SetTheme(m.config.Theme)
				m.settings.SetAIProvider(m.config.AI.Provider)
				m.settings.SetAPIKey(m.config.AI.APIKey)
				m.settings.SetModel(m.config.AI.Model)
				m.settings.Show()
				m.state = StateSettings
			}
		case components.ActionDelete:
			// Delete connection
			if connIdx >= 0 && connIdx < len(m.config.Connections) {
				connName := m.config.Connections[connIdx].Name
				m.config.RemoveConnection(connIdx)
				m.loadConnections()
				m.config.Save()
				m.statusMessage = "Deleted connection: " + connName
				m.isError = false
			}
		case components.ActionCancel:
			// Do nothing
		}
		return m, nil
	}
	return m, nil
}

// updateSettings handles settings modal state
func (m *Model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.settings.Hide()
		m.state = StateNormal
		return m, nil
	case "ctrl+s", "f5":
		// Test connection without saving
		name, driver, host, port, user, pass, database := m.settings.GetConnectionConfig()
		if name != "" && host != "" {
			testCfg := &config.DatabaseConfig{
				Name:     name,
				Driver:   driver,
				Host:     host,
				Port:     parsePort(port, driver),
				User:     user,
				Password: pass,
				Database: database,
			}
			m.settings.SetStatus("Testing connection...", false)
			if err := m.TestConnection(testCfg); err != nil {
				m.settings.SetStatus("❌ Test failed: "+err.Error(), true)
			} else {
				m.settings.SetStatus("✅ Connection test successful!", false)
			}
		} else {
			m.settings.SetStatus("❌ Please fill in name and host", true)
		}
		return m, nil
	case "enter":
		// Validate connection if on Connections tab
		name, driver, host, port, user, pass, database := m.settings.GetConnectionConfig()
		if name != "" && host != "" {
			// Test connection before saving
			testConn := config.DatabaseConfig{
				Name:     name,
				Driver:   driver,
				Host:     host,
				Port:     parsePort(port, driver),
				User:     user,
				Password: pass,
				Database: database,
			}
			m.settings.SetStatus("Validating connection...", false)
			if err := m.TestConnection(&testConn); err != nil {
				m.settings.SetStatus("❌ ERROR: "+err.Error(), true)
				return m, nil
			}
			m.settings.SetStatus("✅ CONNECTED!", false)
		}
		
		// Apply settings
		theme := m.settings.GetSelectedTheme()
		if theme != m.config.Theme {
			m.UpdateTheme(theme)
			m.config.Theme = theme
		}
		
		m.config.AI.Provider = m.settings.GetSelectedProvider()
		m.config.AI.APIKey = m.settings.GetAPIKey()
		m.config.AI.Model = m.settings.GetModel()
		
		// Reinitialize AI provider
		if m.config.AI.Provider != "none" {
			provider, _ := NewAIProvider(m.config.AI.Provider, m.config.AI.APIKey, m.config.AI.Model)
			m.aiProvider = provider
		}
		
		// Handle connection from Connections tab
		if name != "" && host != "" {
			newConn := config.DatabaseConfig{
				Name:     name,
				Driver:   driver,
				Host:     host,
				Port:     parsePort(port, driver),
				User:     user,
				Password: pass,
				Database: database,
			}
			
			if m.settings.IsEditingConnection() {
				// Update existing connection
				editIdx := m.settings.GetEditingConnIndex()
				if editIdx >= 0 && editIdx < len(m.config.Connections) {
					m.config.Connections[editIdx] = newConn
					m.statusMessage = "Connection updated: " + name
					m.isError = false
					
					// If updating active connection, reconnect
					if m.config.ActiveConnIndex == editIdx {
						if err := m.Connect(); err != nil {
							m.statusMessage = "Updated but failed to connect: " + err.Error()
							m.isError = true
						}
					}
				}
			} else {
				// Add new connection
				m.config.Connections = append(m.config.Connections, newConn)
				m.config.ActiveConnIndex = len(m.config.Connections) - 1
				
				// Connect to the new database
				if err := m.Connect(); err != nil {
					m.statusMessage = "Added but failed to connect: " + err.Error()
					m.isError = true
				} else {
					m.statusMessage = "Connection added: " + name
					m.isError = false
				}
			}
			m.loadConnections()
		}
		
		m.config.Save()
		m.settings.Hide()
		m.state = StateNormal
		if m.statusMessage == "" {
			m.statusMessage = "Settings saved"
			m.isError = false
		}
		return m, nil
	default:
		var cmd tea.Cmd
		m.settings, cmd = m.settings.Update(msg)
		return m, cmd
	}
}

// parsePort parses port string to int with defaults
func parsePort(portStr string, driver string) int {
	if portStr == "" {
		switch driver {
		case "postgres":
			return 5432
		case "mysql":
			return 3306
		default:
			return 0
		}
	}
	port := 0
	for _, c := range portStr {
		if c >= '0' && c <= '9' {
			port = port*10 + int(c-'0')
		}
	}
	return port
}

// updateNormal handles normal operation state
func (m *Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global shortcuts (always work regardless of focused pane)
	switch key {
	case "f5", "ctrl+e":
		m.ExecuteQuery()
		return m, nil

	case "ctrl+g":
		// Set context if there's a selection
		selectedText := m.editor.GetSelectedText()
		if selectedText != m.editor.GetValue() {
			m.aiPrompt.SetContext(selectedText)
		} else {
			m.aiPrompt.ClearContext()
		}
		m.aiPrompt.Show(components.AIPromptModeNL2SQL)
		m.state = StateAIPrompt
		return m, nil

	case "ctrl+k":
		// Set context if there's a selection
		selectedText := m.editor.GetSelectedText()
		if selectedText != m.editor.GetValue() {
			m.aiPrompt.SetContext(selectedText)
		} else {
			m.aiPrompt.ClearContext()
		}
		m.aiPrompt.Show(components.AIPromptModeRefactor)
		m.state = StateAIPrompt
		return m, nil
		
	case "f1":
		m.FocusNext()
		return m, nil

	case "f2":
		m.FocusPrev()
		return m, nil
	
	case "f3":
		// Toggle completion panel (show/hide)
		if m.completion.IsVisible() {
			m.completion.Hide()
			m.statusMessage = "Keywords panel hidden"
		} else {
			m.triggerCompletion()
			m.statusMessage = "Keywords panel shown"
		}
		m.isError = false
		return m, nil
	
	case "f4":
		// Toggle Help modal
		m.help.Toggle()
		return m, nil
	}
	
	// Handle Help modal navigation when visible
	if m.help.IsVisible() {
		switch msg.String() {
		case "left", "h":
			m.help.PrevPage()
			return m, nil
		case "right", "l":
			m.help.NextPage()
			return m, nil
		case "esc", "q":
			m.help.Hide()
			return m, nil
		}
	}

	// Handle completion popup navigation (only when visible and editor focused)
	if m.completion.IsVisible() && m.focusedPane == PaneEditor {
		switch msg.String() {
		case "up", "ctrl+p":
			m.completion.MoveUp()
			return m, nil
		case "down", "ctrl+n":
			m.completion.MoveDown()
			return m, nil
		case "tab":
			// Accept completion - replace current word with suggestion
			item := m.completion.GetSelected()
			if item != nil {
				m.editor.ReplaceCurrentWord(item.InsertText)
			}
			// Don't hide - keep panel open
			return m, nil
		}
	}

	// Handle pane-specific keys
	switch m.focusedPane {
	case PaneSidebar:
		return m.handleSidebarKeys(msg)
	case PaneEditor:
		// Pass all keys to editor (for typing)
		var cmd tea.Cmd
		m.editor, cmd = m.editor.Update(msg)
		
		// Auto-refresh completions if panel is visible (on change)
		if m.completion.IsVisible() {
			m.refreshCompletions()
		}
		
		return m, cmd
	case PaneResults:
		return m.handleResultsKeys(msg)
	}

	return m, nil
}

// handleSidebarKeys handles keys when sidebar is focused
func (m *Model) handleSidebarKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left":
		// Cycle sections: Connections -> Databases -> Tables -> Connections
		section := m.sidebar.GetSection()
		if section == components.SectionConnections {
			m.sidebar.SetSection(components.SectionTables)
		} else if section == components.SectionDatabases {
			m.sidebar.SetSection(components.SectionConnections)
		} else {
			m.sidebar.SetSection(components.SectionDatabases)
		}
		return m, nil
	case "right":
		// Cycle sections: Connections -> Databases -> Tables -> Connections
		section := m.sidebar.GetSection()
		if section == components.SectionConnections {
			m.sidebar.SetSection(components.SectionDatabases)
		} else if section == components.SectionDatabases {
			m.sidebar.SetSection(components.SectionTables)
		} else {
			m.sidebar.SetSection(components.SectionConnections)
		}
		return m, nil

	case "enter":
		section := m.sidebar.GetSection()
		
		// Handle Connections section
		if section == components.SectionConnections {
			if m.sidebar.IsAddConnectionSelected() {
				// Show settings with Connections tab and clear form
				m.settings.SetTheme(m.config.Theme)
				m.settings.SetAIProvider(m.config.AI.Provider)
				m.settings.SetAPIKey(m.config.AI.APIKey)
				m.settings.SetModel(m.config.AI.Model)
				m.settings.ShowForConnection()
				m.state = StateSettings
				return m, nil
			}
			connIdx := m.sidebar.GetSelectedConnection()
			if connIdx >= 0 && connIdx < len(m.config.Connections) {
				// Show connection action modal
				connName := m.config.Connections[connIdx].Name
				m.connModal.Show(connName, connIdx)
				m.state = StateConnModal
			}
			return m, nil
		}
		
		// Handle Databases section
		if section == components.SectionDatabases {
			dbName := m.sidebar.GetSelectedDatabase()
			if dbName != "" && dbName != m.sidebar.GetCurrentDatabase() {
				if err := m.SwitchDatabase(dbName); err != nil {
					m.statusMessage = "Failed to switch: " + err.Error()
					m.isError = true
				}
			}
			return m, nil
		}
		
		// Handle Tables section
		if section == components.SectionTables {
			tableName := m.sidebar.SelectedTable()
			if tableName != "" {
				// Mark table as selected
				m.sidebar.SelectTable(tableName)
				// Save to config
				m.config.LastTable = tableName
				m.config.Save()
				// Insert SELECT * query
				query := fmt.Sprintf("SELECT * FROM %s LIMIT 100;", tableName)
				m.editor.SetValue(query)
				m.FocusEditor()
				m.statusMessage = "Selected table: " + tableName
				m.isError = false
			}
			return m, nil
		}
	}
	
	// Pass other keys to sidebar
	var cmd tea.Cmd
	m.sidebar, cmd = m.sidebar.Update(msg)
	return m, cmd
}

// Helper to get SectionConnections
func SectionConnections() components.SidebarSection {
	return components.SectionConnections
}

// handleResultsKeys handles keys when results pane is focused
func (m *Model) handleResultsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "pgdown", "ctrl+d":
		m.results.NextPage()
		return m, nil
	case "pgup", "ctrl+u":
		m.results.PrevPage()
		return m, nil
	case "c":
		// Copy selected row
		if err := m.results.CopySelectedRow(); err != nil {
			m.statusMessage = "Copy failed: " + err.Error()
			m.isError = true
		} else {
			m.statusMessage = "Row copied to clipboard"
			m.isError = false
		}
		return m, nil
	case "C", "ctrl+shift+c":
		// Copy all data
		if err := m.results.CopyAllData(); err != nil {
			m.statusMessage = "Copy failed: " + err.Error()
			m.isError = true
		} else {
			m.statusMessage = fmt.Sprintf("All data (%d rows) copied to clipboard", m.results.GetRowCount())
			m.isError = false
		}
		return m, nil
	case "v":
		// Cycle view modes
		current := m.results.GetViewMode()
		next := (current + 1) % 4 // 4 modes
		m.results.SetViewMode(next)
		return m, nil
	case "1":
		m.results.SetViewMode(components.ViewTable)
		return m, nil
	case "2":
		m.results.SetViewMode(components.ViewChartBar)
		return m, nil
	case "3":
		m.results.SetViewMode(components.ViewChartLine)
		return m, nil
	case "4":
		m.results.SetViewMode(components.ViewChartPie)
		return m, nil
	}

	var cmd tea.Cmd
	m.results, cmd = m.results.Update(msg)
	return m, cmd
}

// updateLayout updates component sizes based on window size
func (m *Model) updateLayout() {
	sidebarWidth := 20
	if m.width < 80 {
		sidebarWidth = 15
	}
	
	mainWidth := m.width - sidebarWidth - 2
	headerHeight := 1
	footerHeight := 1
	contentHeight := m.height - headerHeight - footerHeight - 2
	
	editorHeight := contentHeight * 40 / 100
	resultsHeight := contentHeight - editorHeight

	m.sidebar.SetSize(sidebarWidth, contentHeight)
	m.editor.SetSize(mainWidth, editorHeight)
	m.editor.SetPosition(sidebarWidth, headerHeight+1) // +1 for newline
	m.results.SetSize(mainWidth, resultsHeight)
	
	modalWidth := m.width * 60 / 100
	if modalWidth < 50 {
		modalWidth = 50
	}
	m.aiPrompt.SetSize(modalWidth, 10)
	m.settings.SetSize(modalWidth, m.height*70/100)
	m.wizard.SetSize(m.width, m.height)
}

// NewAIProvider is a helper to create AI providers
func NewAIProvider(provider, apiKey, model string) (ai.Provider, error) {
	return ai.NewProvider(provider, apiKey, model)
}
