package tui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/febritecno/sqdesk/internal/ai"
	"github.com/febritecno/sqdesk/internal/config"
	"github.com/febritecno/sqdesk/internal/db"
	"github.com/febritecno/sqdesk/internal/tui/components"
	"github.com/febritecno/sqdesk/internal/tui/setup"
)

// Pane represents the focused pane
type Pane int

const (
	PaneSidebar Pane = iota
	PaneEditor
	PaneResults
)

// AppState represents the application state
type AppState int

const (
	StateSetup AppState = iota
	StateNormal
	StateAIPrompt
	StateSettings
)

// Model is the main application model
type Model struct {
	// Configuration
	config *config.Config
	styles *Styles

	// Database
	connector db.Connector
	schema    *db.Schema
	tables    []string

	// AI
	aiProvider ai.Provider

	// UI Components
	sidebar   components.Sidebar
	editor    components.Editor
	results   components.Results
	aiPrompt  components.AIPrompt
	settings  components.Settings
	wizard    *setup.Wizard

	// State
	state       AppState
	focusedPane Pane
	width       int
	height      int
	
	// Status
	statusMessage string
	isError       bool
	isConnected   bool
	
	// Query
	lastQuery     string
	queryRunning  bool
}

// NewModel creates a new application model
func NewModel(cfg *config.Config) *Model {
	// Get theme colors
	colors := GetThemeColors(cfg.Theme)
	styles := NewStyles(colors)

	// Create component styles
	sidebarStyles := components.SidebarStyles{
		Normal:       styles.Panel,
		Focused:      styles.PanelFocused,
		Title:        styles.PanelTitle,
		Item:         styles.SidebarItem,
		Selected:     styles.SidebarSelected,
		AddButton:    styles.AddButton,
		ActiveConn:   styles.ActiveConn,
		InactiveConn: styles.InactiveConn,
	}

	editorStyles := components.EditorStyles{
		Normal:    styles.Panel,
		Focused:   styles.PanelFocused,
		Title:     styles.PanelTitle,
		LineNum:   styles.InfoText,
		Keyword:   styles.Keyword,
		String:    styles.String,
		GhostText: styles.GhostText,
		Suggestion: styles.InfoText.Copy().Foreground(lipgloss.Color("208")).Bold(true),
		Type:       styles.Keyword.Copy().Foreground(lipgloss.Color("33")), // Blue
		Function:   styles.Keyword.Copy().Foreground(lipgloss.Color("220")), // Yellow
		Operator:   styles.Keyword.Copy().Foreground(lipgloss.Color("201")), // Pink
	}

	resultsStyles := components.ResultsStyles{
		Normal:      styles.Panel,
		Focused:     styles.PanelFocused,
		Title:       styles.PanelTitle,
		Header:      styles.ResultsHeader,
		Cell:        styles.ResultsCell,
		SelectedRow: styles.SidebarSelected,
		Error:       styles.ErrorText,
		Info:        styles.InfoText,
	}

	aiPromptStyles := components.AIPromptStyles{
		Modal:        styles.Modal,
		Title:        styles.ModalTitle,
		Input:        styles.Input,
		Hint:         styles.HelpDesc,
		ModeNL2SQL:   styles.InfoText,
		ModeRefactor: styles.WarningText,
	}

	settingsStyles := components.SettingsStyles{
		Modal:        styles.Modal,
		Title:        styles.ModalTitle,
		Tab:          styles.Button,
		TabActive:    styles.ButtonActive,
		Label:        styles.InputLabel,
		Input:        styles.Input,
		InputFocus:   styles.InputFocused,
		Button:       styles.Button,
		ButtonActive: styles.ButtonActive,
		Selected:     styles.SidebarSelected,
		Hint:         styles.HelpDesc,
	}

	wizardStyles := setup.WizardStyles{
		Container:    styles.App,
		Title:        styles.ModalTitle,
		Subtitle:     styles.HelpDesc,
		ASCII:        styles.InfoText,
		Text:         styles.ModalContent,
		Selected:     styles.ButtonActive,
		Unselected:   styles.Button,
		Input:        styles.Input,
		InputFocus:   styles.InputFocused,
		Label:        styles.InputLabel,
		Button:       styles.Button,
		ButtonActive: styles.ButtonActive,
		Hint:         styles.HelpDesc,
		Success:      styles.SuccessText,
	}

	// Determine initial state
	state := StateNormal
	if cfg.FirstRun {
		state = StateSetup
	}

	m := &Model{
		config:      cfg,
		styles:      styles,
		state:       state,
		focusedPane: PaneSidebar,
		sidebar:     components.NewSidebar(sidebarStyles),
		editor:      components.NewEditor(editorStyles),
		results:     components.NewResults(resultsStyles),
		aiPrompt:    components.NewAIPrompt(aiPromptStyles),
		settings:    components.NewSettings(settingsStyles),
		wizard:      setup.NewWizard(wizardStyles),
	}

	// Initialize AI provider if configured
	if cfg.AI.Provider != "none" && cfg.AI.Provider != "" {
		provider, _ := ai.NewProvider(cfg.AI.Provider, cfg.AI.APIKey, cfg.AI.Model)
		m.aiProvider = provider
	} else {
		m.aiProvider = ai.NewNoopProvider()
	}

	// Load connections into sidebar
	m.loadConnections()

	// Set focus
	m.sidebar.SetFocused(true)
	m.editor.SetFocused(false)
	m.results.SetFocused(false)

	return m
}

// loadConnections loads connections from config into sidebar
func (m *Model) loadConnections() {
	conns := make([]components.ConnectionItem, len(m.config.Connections))
	for i, conn := range m.config.Connections {
		conns[i] = components.ConnectionItem{
			Name:   conn.Name,
			Driver: conn.Driver,
			Active: i == m.config.ActiveConnIndex,
		}
	}
	m.sidebar.SetConnections(conns)
}

// Connect connects to the active database
func (m *Model) Connect() error {
	connCfg := m.config.GetActiveConnection()
	if connCfg == nil {
		return fmt.Errorf("no database connection configured")
	}

	// Close existing connection if any
	if m.connector != nil {
		m.connector.Close()
		m.connector = nil
		m.isConnected = false
	}

	connector, err := db.NewConnector(connCfg)
	if err != nil {
		return fmt.Errorf("failed to create connector: %w", err)
	}

	if err := connector.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Verify connection is working
	if !connector.IsConnected() {
		return fmt.Errorf("connection test failed")
	}

	m.connector = connector
	m.isConnected = true

	// Load tables
	tables, err := connector.GetTables()
	if err != nil {
		m.statusMessage = "Connected, but failed to load tables: " + err.Error()
		m.isError = true
	} else {
		m.tables = tables
		m.sidebar.SetTables(tables)
		m.statusMessage = fmt.Sprintf("Connected to %s", connCfg.Name)
		m.isError = false
	}

	// Load schema for auto-completion
	schema, err := connector.GetSchema()
	if err == nil {
		m.schema = schema
		// Convert schema to map for editor
		schemaMap := make(map[string][]string)
		for tableName, table := range schema.Tables {
			cols := make([]string, len(table.Columns))
			for i, col := range table.Columns {
				cols[i] = col.Name
			}
			schemaMap[tableName] = cols
		}
		m.editor.SetSchema(schemaMap)
	}

	// Load available databases
	m.LoadDatabases()

	return nil
}

// LoadDatabases loads the list of available databases
func (m *Model) LoadDatabases() {
	if m.connector == nil {
		return
	}
	
	databases, err := m.connector.GetDatabases()
	if err != nil {
		return
	}
	
	currentDB := m.connector.GetDatabaseName()
	m.sidebar.SetDatabases(databases, currentDB)
}

// SwitchDatabase switches to a different database
func (m *Model) SwitchDatabase(dbName string) error {
	if m.connector == nil {
		return fmt.Errorf("not connected")
	}
	
	if err := m.connector.SwitchDatabase(dbName); err != nil {
		return err
	}
	
	// Reload tables for new database
	tables, err := m.connector.GetTables()
	if err == nil {
		m.tables = tables
		m.sidebar.SetTables(tables)
	}
	
	// Reload schema
	schema, err := m.connector.GetSchema()
	if err == nil {
		m.schema = schema
		schemaMap := make(map[string][]string)
		for tableName, table := range schema.Tables {
			cols := make([]string, len(table.Columns))
			for i, col := range table.Columns {
				cols[i] = col.Name
			}
			schemaMap[tableName] = cols
		}
		m.editor.SetSchema(schemaMap)
	}
	
	// Update sidebar
	m.LoadDatabases()
	
	m.statusMessage = "Switched to database: " + dbName
	m.isError = false
	return nil
}

// TestConnection tests a database connection without storing it
func (m *Model) TestConnection(cfg *config.DatabaseConfig) error {
	connector, err := db.NewConnector(cfg)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	if err := connector.Connect(); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	
	// Test the connection
	if !connector.IsConnected() {
		connector.Close()
		return fmt.Errorf("connection test failed")
	}

	connector.Close()
	return nil
}

// Disconnect disconnects from the current database
func (m *Model) Disconnect() {
	if m.connector != nil {
		m.connector.Close()
		m.connector = nil
	}
	m.isConnected = false
	m.tables = nil
	m.sidebar.SetTables(nil)
	m.schema = nil
	m.statusMessage = "Disconnected"
	m.isError = false
}

// ExecuteQuery executes the current SQL query
func (m *Model) ExecuteQuery() {
	// Use selected text if available, otherwise full content
	sql := m.editor.GetSelectedText()
	if sql == "" {
		sql = m.editor.GetValue()
	}

	if strings.TrimSpace(sql) == "" {
		m.results.SetMessage("No query to execute")
		return
	}

	if m.connector == nil || !m.isConnected {
		m.results.SetError(fmt.Errorf("not connected to database"))
		return
	}

	m.queryRunning = true
	m.lastQuery = sql

	// Determine query type
	// Strip comments and leading whitespace to find the actual command
	cleanSQL := sql
	// Strip single line comments
	reLine := regexp.MustCompile(`(?m)^--.*$`)
	cleanSQL = reLine.ReplaceAllString(cleanSQL, "")
	// Strip multi-line comments
	reMulti := regexp.MustCompile(`(?s)/\*.*?\*/`)
	cleanSQL = reMulti.ReplaceAllString(cleanSQL, "")
	
	trimmedSQL := strings.TrimSpace(strings.ToLower(cleanSQL))
	isSelect := strings.HasPrefix(trimmedSQL, "select") ||
				strings.HasPrefix(trimmedSQL, "show") ||
				strings.HasPrefix(trimmedSQL, "describe") ||
				strings.HasPrefix(trimmedSQL, "explain") ||
				strings.HasPrefix(trimmedSQL, "with") ||
				strings.HasPrefix(trimmedSQL, "pragma") // SQLite pragma often returns data

	if isSelect {
		rows, columns, err := m.connector.Query(sql)
		if err != nil {
			m.results.SetError(err)
			m.statusMessage = "Query failed"
			m.isError = true
		} else {
			m.results.SetData(columns, rows)
			m.statusMessage = fmt.Sprintf("Query returned %d rows", len(rows))
			m.isError = false
			// Default to table view for new results
			m.results.SetViewMode(components.ViewTable)
		}
	} else {
		// Execute non-select query
		affected, err := m.connector.Execute(sql)
		if err != nil {
			m.results.SetError(err)
			m.statusMessage = "Execution failed"
			m.isError = true
		} else {
			m.results.SetMessage(fmt.Sprintf("Query executed successfully. %d rows affected.", affected))
			m.statusMessage = fmt.Sprintf("Affected %d rows", affected)
			m.isError = false
		}
	}

	m.queryRunning = false
}

// PreviewTable previews the selected table
func (m *Model) PreviewTable(tableName string) {
	if tableName == "" {
		return
	}

	sql := fmt.Sprintf("SELECT * FROM %s LIMIT 100", tableName)
	m.editor.SetValue(sql)
	m.ExecuteQuery()
}

// GenerateSQL uses AI to generate SQL from natural language
func (m *Model) GenerateSQL(prompt string) {
	if m.aiProvider == nil || !m.aiProvider.IsConfigured() {
		m.statusMessage = "AI not configured"
		m.isError = true
		return
	}

	sql, err := m.aiProvider.NL2SQL(prompt, m.schema)
	if err != nil {
		m.statusMessage = "AI error: " + err.Error()
		m.isError = true
		return
	}

	m.editor.SetValue(sql)
	m.statusMessage = "SQL generated by AI"
	m.isError = false
}

// RefactorSQL uses AI to refactor SQL
func (m *Model) RefactorSQL(instruction string) {
	if m.aiProvider == nil || !m.aiProvider.IsConfigured() {
		m.statusMessage = "AI not configured"
		m.isError = true
		return
	}

	currentSQL := m.editor.GetValue()
	if currentSQL == "" {
		m.statusMessage = "No SQL to refactor"
		m.isError = true
		return
	}

	sql, err := m.aiProvider.RefactorSQL(currentSQL, instruction, m.schema)
	if err != nil {
		m.statusMessage = "AI error: " + err.Error()
		m.isError = true
		return
	}

	m.editor.SetValue(sql)
	m.statusMessage = "SQL refactored by AI"
	m.isError = false
}

// FocusNext moves focus to the next pane
func (m *Model) FocusNext() {
	m.sidebar.SetFocused(false)
	m.editor.SetFocused(false)
	m.results.SetFocused(false)

	m.focusedPane = (m.focusedPane + 1) % 3

	switch m.focusedPane {
	case PaneSidebar:
		m.sidebar.SetFocused(true)
	case PaneEditor:
		m.editor.SetFocused(true)
	case PaneResults:
		m.results.SetFocused(true)
	}
}

// FocusPrev moves focus to the previous pane
func (m *Model) FocusPrev() {
	m.sidebar.SetFocused(false)
	m.editor.SetFocused(false)
	m.results.SetFocused(false)

	if m.focusedPane == 0 {
		m.focusedPane = 2
	} else {
		m.focusedPane--
	}

	switch m.focusedPane {
	case PaneSidebar:
		m.sidebar.SetFocused(true)
	case PaneEditor:
		m.editor.SetFocused(true)
	case PaneResults:
		m.results.SetFocused(true)
	}
}

// UpdateTheme updates the theme and regenerates styles
func (m *Model) UpdateTheme(themeName string) {
	m.config.Theme = themeName
	colors := GetThemeColors(themeName)
	m.styles = NewStyles(colors)
}

// GetConnectionInfo returns the current connection info string
func (m *Model) GetConnectionInfo() string {
	if !m.isConnected || m.connector == nil {
		return "Not Connected"
	}
	connCfg := m.config.GetActiveConnection()
	if connCfg != nil {
		return connCfg.Name
	}
	return "Connected"
}

// GetAIInfo returns the AI provider info string
func (m *Model) GetAIInfo() string {
	if m.aiProvider == nil || !m.aiProvider.IsConfigured() {
		return "AI: Disabled"
	}
	return fmt.Sprintf("%s: %s", m.aiProvider.GetProviderName(), m.aiProvider.GetModelName())
}

// Close cleans up resources
func (m *Model) Close() error {
	if m.connector != nil {
		return m.connector.Close()
	}
	return nil
}
