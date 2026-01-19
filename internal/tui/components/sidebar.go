package components

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SidebarSection represents which section is focused
type SidebarSection int

const (
	SectionConnections SidebarSection = iota
	SectionDatabases
	SectionTables
)

// ConnectionItem represents a connection in the sidebar
type ConnectionItem struct {
	Name   string
	Driver string
	Active bool
}

func (c ConnectionItem) Title() string       { return c.Name }
func (c ConnectionItem) Description() string { return c.Driver }
func (c ConnectionItem) FilterValue() string { return c.Name }

// DatabaseItem represents a database in the sidebar
type DatabaseItem string

func (d DatabaseItem) Title() string       { return string(d) }
func (d DatabaseItem) Description() string { return "" }
func (d DatabaseItem) FilterValue() string { return string(d) }

// TableItem represents a table in the sidebar
type TableItem struct {
	name     string
	selected bool
}

func (t TableItem) Title() string       { return t.name }
func (t TableItem) Description() string { return "" }
func (t TableItem) FilterValue() string { return t.name }

// Sidebar component for displaying connections, databases and tables
type Sidebar struct {
	connList      list.Model
	dbList        list.Model
	tableList     list.Model
	
	currentDB     string
	width         int
	height        int
	focused       bool
	section       SidebarSection
	styles        SidebarStyles
}

// SidebarStyles holds styling for the sidebar
type SidebarStyles struct {
	Normal      lipgloss.Style
	Focused     lipgloss.Style
	Title       lipgloss.Style
	Item        lipgloss.Style
	Selected    lipgloss.Style
	AddButton   lipgloss.Style
	ActiveConn  lipgloss.Style
	InactiveConn lipgloss.Style
}

// NewSidebar creates a new sidebar component
func NewSidebar(styles SidebarStyles) Sidebar {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetSpacing(0)
	
	// Connections List
	connList := list.New([]list.Item{}, delegate, 20, 5)
	connList.Title = "CONNECTIONS"
	connList.SetShowStatusBar(false)
	connList.SetShowHelp(false)
	connList.SetFilteringEnabled(false)
	connList.Styles.Title = styles.Title
	
	// Databases List
	dbList := list.New([]list.Item{}, delegate, 20, 5)
	dbList.Title = "DATABASES"
	dbList.SetShowStatusBar(false)
	dbList.SetShowHelp(false)
	dbList.SetFilteringEnabled(true)
	dbList.Styles.Title = styles.Title
	
	// Tables List
	tableList := list.New([]list.Item{}, delegate, 20, 10)
	tableList.Title = "TABLES"
	tableList.SetShowStatusBar(false)
	tableList.SetShowHelp(false)
	tableList.SetFilteringEnabled(true)
	tableList.Styles.Title = styles.Title
	
	return Sidebar{
		connList:  connList,
		dbList:    dbList,
		tableList: tableList,
		focused:   false,
		section:   SectionConnections,
		styles:    styles,
	}
}

// AddConnectionItem represents the add button
type AddConnectionItem struct{}

func (a AddConnectionItem) Title() string       { return "+ Add Connection" }
func (a AddConnectionItem) Description() string { return "Create a new connection" }
func (a AddConnectionItem) FilterValue() string { return "add connection" }

// SetConnections sets the list of connections
func (s *Sidebar) SetConnections(connections []ConnectionItem) {
	items := make([]list.Item, len(connections)+1)
	for i, c := range connections {
		items[i] = c
	}
	items[len(connections)] = AddConnectionItem{}
	s.connList.SetItems(items)
}

// AddConnection adds a connection to the list
func (s *Sidebar) AddConnection(conn ConnectionItem) {
	s.connList.InsertItem(len(s.connList.Items()), conn)
}

// SetActiveConnection sets which connection is active
func (s *Sidebar) SetActiveConnection(index int) {
	items := s.connList.Items()
	for i, item := range items {
		conn := item.(ConnectionItem)
		conn.Active = (i == index)
		items[i] = conn
	}
	s.connList.SetItems(items)
}

// GetSelectedConnection returns the selected connection index
func (s Sidebar) GetSelectedConnection() int {
	return s.connList.Index()
}

// SetTables sets the list of tables
func (s *Sidebar) SetTables(tables []string) {
	items := make([]list.Item, len(tables))
	for i, t := range tables {
		items[i] = TableItem{name: t}
	}
	s.tableList.SetItems(items)
}

// SetDatabases sets the list of databases
func (s *Sidebar) SetDatabases(databases []string, current string) {
	s.currentDB = current
	items := make([]list.Item, len(databases))
	for i, d := range databases {
		items[i] = DatabaseItem(d)
		if d == current {
			s.dbList.Select(i)
		}
	}
	s.dbList.SetItems(items)
}

// GetSelectedDatabase returns the selected database name
func (s Sidebar) GetSelectedDatabase() string {
	if s.dbList.SelectedItem() == nil {
		return ""
	}
	return string(s.dbList.SelectedItem().(DatabaseItem))
}

// GetCurrentDatabase returns the current active database
func (s Sidebar) GetCurrentDatabase() string {
	return s.currentDB
}

// SetSize sets the sidebar dimensions
func (s *Sidebar) SetSize(width, height int) {
	s.width = width
	s.height = height
    
    // Distribute height
    minHeight := 5
    
    connHeight := height / 4
    if connHeight < minHeight { connHeight = minHeight }
    
    dbHeight := height / 4
    if dbHeight < minHeight { dbHeight = minHeight }
    
    tableHeight := height - connHeight - dbHeight
    if tableHeight < minHeight { tableHeight = minHeight }
    
	s.connList.SetSize(width-2, connHeight)
    s.dbList.SetSize(width-2, dbHeight)
    s.tableList.SetSize(width-2, tableHeight)
}

// SetFocused sets the focus state
func (s *Sidebar) SetFocused(focused bool) {
	s.focused = focused
}

// IsFocused returns if sidebar is focused
func (s Sidebar) IsFocused() bool {
	return s.focused
}

// GetSection returns the current section
func (s Sidebar) GetSection() SidebarSection {
	return s.section
}

// SetSection sets the current section
func (s *Sidebar) SetSection(section SidebarSection) {
	s.section = section
}

// ToggleSection switches between connections and tables
func (s *Sidebar) ToggleSection() {
	if s.section == SectionConnections {
		s.section = SectionDatabases
	} else if s.section == SectionDatabases {
        s.section = SectionTables
    } else {
		s.section = SectionConnections
	}
}

// SelectedTable returns the currently selected table name
func (s Sidebar) SelectedTable() string {
	if item := s.tableList.SelectedItem(); item != nil {
		return item.(TableItem).name
	}
	return ""
}

// IsAddConnectionSelected returns true if "+ Add Connection" is selected
func (s Sidebar) IsAddConnectionSelected() bool {
	if s.section != SectionConnections { return false }
    item := s.connList.SelectedItem()
    if item == nil { return false }
    _, ok := item.(AddConnectionItem)
    return ok
}

// Update handles input for the sidebar
func (s Sidebar) Update(msg tea.Msg) (Sidebar, tea.Cmd) {
	if !s.focused {
		return s, nil
	}

	var cmd tea.Cmd
    var cmds []tea.Cmd
    
    // Only update the active list
    switch s.section {
    case SectionConnections:
        s.connList, cmd = s.connList.Update(msg)
        cmds = append(cmds, cmd)
    case SectionDatabases:
        s.dbList, cmd = s.dbList.Update(msg)
        cmds = append(cmds, cmd)
    case SectionTables:
        s.tableList, cmd = s.tableList.Update(msg)
        cmds = append(cmds, cmd)
    }
    
	return s, tea.Batch(cmds...)
}

// View renders the sidebar
func (s Sidebar) View() string {
	style := s.styles.Normal
	if s.focused {
		style = s.styles.Focused
	}
    
    // Update titles to show active section
    connTitle := "CONNECTIONS"
    dbTitle := "DATABASES"
    tableTitle := "TABLES"
    
    if s.section == SectionConnections { connTitle = "▼ " + connTitle } else { connTitle = "▶ " + connTitle }
    if s.section == SectionDatabases { dbTitle = "▼ " + dbTitle } else { dbTitle = "▶ " + dbTitle }
    if s.section == SectionTables { tableTitle = "▼ " + tableTitle } else { tableTitle = "▶ " + tableTitle }
    
    s.connList.Title = connTitle
    s.dbList.Title = dbTitle
    s.tableList.Title = tableTitle
    
    return style.
        Width(s.width).
        Height(s.height).
        Render(lipgloss.JoinVertical(lipgloss.Left,
            s.connList.View(),
            s.dbList.View(),
            s.tableList.View(),
        ))
}

// Helper function
func truncate(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	if maxLen <= 3 {
		return str[:maxLen]
	}
	return str[:maxLen-3] + "..."
}

// GetConnections returns the connections list
func (s Sidebar) GetConnections() []ConnectionItem {
    items := s.connList.Items()
    conns := make([]ConnectionItem, 0, len(items))
    for _, item := range items {
        if c, ok := item.(ConnectionItem); ok {
            conns = append(conns, c)
        }
    }
    return conns
}

// GetDatabases returns the databases list
func (s Sidebar) GetDatabases() []string {
    items := s.dbList.Items()
    dbs := make([]string, len(items))
    for i, item := range items {
        dbs[i] = string(item.(DatabaseItem))
    }
    return dbs
}
