package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/febritecno/sqdesk/internal/ai"
	"github.com/febritecno/sqdesk/internal/config"
	"github.com/febritecno/sqdesk/internal/tui/components"
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
		m.state = StateNormal
		return m, nil
	default:
		var cmd tea.Cmd
		m.aiPrompt, cmd = m.aiPrompt.Update(msg)
		return m, cmd
	}
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
			if err := m.TestConnection(testCfg); err != nil {
				m.statusMessage = "Test failed: " + err.Error()
				m.isError = true
			} else {
				m.statusMessage = "Connection test successful!"
				m.isError = false
			}
		}
		return m, nil
	case "enter":
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
		
		// Handle new connection from Connections tab
		name, driver, host, port, user, pass, database := m.settings.GetConnectionConfig()
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
			
			// Test connection before adding
			if err := m.TestConnection(&newConn); err != nil {
				m.statusMessage = "Connection failed: " + err.Error()
				m.isError = true
				return m, nil
			}
			
			// Add to config
			m.config.Connections = append(m.config.Connections, newConn)
			m.config.ActiveConnIndex = len(m.config.Connections) - 1
			m.loadConnections()
			
			// Connect to the new database
			if err := m.Connect(); err != nil {
				m.statusMessage = "Added but failed to connect: " + err.Error()
				m.isError = true
			} else {
				m.statusMessage = "Connection added: " + name
				m.isError = false
			}
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
		m.aiPrompt.Show(components.AIPromptModeNL2SQL)
		m.state = StateAIPrompt
		return m, nil

	case "ctrl+k":
		m.aiPrompt.Show(components.AIPromptModeRefactor)
		m.state = StateAIPrompt
		return m, nil

	case "f2":
		m.settings.SetTheme(m.config.Theme)
		m.settings.SetAIProvider(m.config.AI.Provider)
		m.settings.SetAPIKey(m.config.AI.APIKey)
		m.settings.SetModel(m.config.AI.Model)
		m.settings.Show()
		m.state = StateSettings
		return m, nil

	case "ctrl+right":
		m.FocusNext()
		return m, nil

	case "ctrl+left":
		m.FocusPrev()
		return m, nil
	}

	// Handle pane-specific keys
	switch m.focusedPane {
	case PaneSidebar:
		return m.handleSidebarKeys(msg)
	case PaneEditor:
		// Pass all keys to editor (for typing)
		var cmd tea.Cmd
		m.editor, cmd = m.editor.Update(msg)
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
				m.settings.SetTheme(m.config.Theme)
				m.settings.SetAIProvider(m.config.AI.Provider)
				m.settings.SetAPIKey(m.config.AI.APIKey)
				m.settings.SetModel(m.config.AI.Model)
				m.settings.Show()
				m.state = StateSettings
				return m, nil
			}
			connIdx := m.sidebar.GetSelectedConnection()
			if connIdx >= 0 && connIdx < len(m.config.Connections) {
				m.config.ActiveConnIndex = connIdx
				m.sidebar.SetActiveConnection(connIdx)
				if m.connector != nil {
					m.connector.Close()
				}
				if err := m.Connect(); err != nil {
					m.statusMessage = "Connection failed: " + err.Error()
					m.isError = true
				} else {
					m.statusMessage = "Connected to " + m.config.Connections[connIdx].Name
					m.isError = false
				}
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
				// Insert SELECT * query
				query := fmt.Sprintf("SELECT * FROM %s LIMIT 100;", tableName)
				m.editor.SetValue(query)
				m.FocusEditor()
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
