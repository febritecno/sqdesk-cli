package tui

import "github.com/charmbracelet/lipgloss"

// Styles holds all TUI styles
type Styles struct {
	// App styles
	App     lipgloss.Style
	Header  lipgloss.Style
	Footer  lipgloss.Style
	
	// Panel styles
	Panel        lipgloss.Style
	PanelFocused lipgloss.Style
	PanelTitle   lipgloss.Style
	
	// Sidebar styles
	Sidebar         lipgloss.Style
	SidebarItem     lipgloss.Style
	SidebarSelected lipgloss.Style
	ActiveConn      lipgloss.Style
	InactiveConn    lipgloss.Style
	AddButton       lipgloss.Style
	
	// Editor styles
	Editor       lipgloss.Style
	EditorCursor lipgloss.Style
	GhostText    lipgloss.Style
	
	// Results table styles
	ResultsHeader lipgloss.Style
	ResultsCell   lipgloss.Style
	ResultsRow    lipgloss.Style
	
	// Status bar styles
	StatusBar   lipgloss.Style
	StatusItem  lipgloss.Style
	StatusError lipgloss.Style
	
	// Modal styles
	Modal        lipgloss.Style
	ModalTitle   lipgloss.Style
	ModalContent lipgloss.Style
	
	// Input styles
	Input        lipgloss.Style
	InputFocused lipgloss.Style
	InputLabel   lipgloss.Style
	
	// Button styles
	Button         lipgloss.Style
	ButtonActive   lipgloss.Style
	ButtonDisabled lipgloss.Style
	
	// Syntax highlighting
	Keyword  lipgloss.Style
	String   lipgloss.Style
	Number   lipgloss.Style
	Comment  lipgloss.Style
	Function lipgloss.Style
	
	// Messages
	ErrorText   lipgloss.Style
	SuccessText lipgloss.Style
	InfoText    lipgloss.Style
	WarningText lipgloss.Style
	
	// Help
	HelpKey  lipgloss.Style
	HelpDesc lipgloss.Style
}

// ThemeColors holds the color palette
type ThemeColors struct {
	Primary          lipgloss.Color
	Secondary        lipgloss.Color
	Accent           lipgloss.Color
	Background       lipgloss.Color
	BackgroundDark   lipgloss.Color
	BackgroundLight  lipgloss.Color
	Text             lipgloss.Color
	TextMuted        lipgloss.Color
	TextBright       lipgloss.Color
	Success          lipgloss.Color
	Warning          lipgloss.Color
	Error            lipgloss.Color
	Info             lipgloss.Color
	Border           lipgloss.Color
	BorderFocus      lipgloss.Color
	Selection        lipgloss.Color
	GhostText        lipgloss.Color
}

// NewStyles creates styles based on theme colors
func NewStyles(colors ThemeColors) *Styles {
	s := &Styles{}
	
	// App styles
	s.App = lipgloss.NewStyle().
		Background(colors.Background)
	
	s.Header = lipgloss.NewStyle().
		Background(colors.BackgroundDark).
		Foreground(colors.Text).
		Padding(0, 1).
		Bold(true)
	
	s.Footer = lipgloss.NewStyle().
		Background(colors.BackgroundDark).
		Foreground(colors.TextMuted).
		Padding(0, 1)
	
	// Panel styles
	s.Panel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colors.Border).
		Background(colors.Background).
		Padding(0, 1)
	
	s.PanelFocused = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colors.BorderFocus).
		Background(colors.Background).
		Padding(0, 1)
	
	s.PanelTitle = lipgloss.NewStyle().
		Foreground(colors.Primary).
		Bold(true).
		Padding(0, 1)
	
	// Sidebar styles
	s.Sidebar = lipgloss.NewStyle().
		Background(colors.BackgroundDark)
	
	s.SidebarItem = lipgloss.NewStyle().
		Foreground(colors.Text).
		Padding(0, 1)
	
	s.SidebarSelected = lipgloss.NewStyle().
		Foreground(colors.TextBright).
		Background(colors.Selection).
		Bold(true).
		Padding(0, 1)
	
	s.ActiveConn = lipgloss.NewStyle().
		Foreground(colors.Success).
		Bold(true).
		Padding(0, 1)
	
	s.InactiveConn = lipgloss.NewStyle().
		Foreground(colors.TextMuted).
		Padding(0, 1)
	
	s.AddButton = lipgloss.NewStyle().
		Foreground(colors.Info).
		Padding(0, 1)
	
	// Editor styles
	s.Editor = lipgloss.NewStyle().
		Foreground(colors.Text)
	
	s.EditorCursor = lipgloss.NewStyle().
		Background(colors.Text).
		Foreground(colors.Background)
	
	s.GhostText = lipgloss.NewStyle().
		Foreground(colors.GhostText).
		Italic(true)
	
	// Results table styles
	s.ResultsHeader = lipgloss.NewStyle().
		Foreground(colors.Primary).
		Bold(true).
		Padding(0, 1)
	
	s.ResultsCell = lipgloss.NewStyle().
		Foreground(colors.Text).
		Padding(0, 1)
	
	s.ResultsRow = lipgloss.NewStyle().
		Background(colors.Background)
	
	// Status bar styles
	s.StatusBar = lipgloss.NewStyle().
		Background(colors.BackgroundDark).
		Foreground(colors.TextMuted).
		Padding(0, 1)
	
	s.StatusItem = lipgloss.NewStyle().
		Foreground(colors.Text).
		Padding(0, 1)
	
	s.StatusError = lipgloss.NewStyle().
		Foreground(colors.Error).
		Bold(true)
	
	// Modal styles
	s.Modal = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colors.Primary).
		Background(colors.BackgroundDark).
		Padding(1, 2)
	
	s.ModalTitle = lipgloss.NewStyle().
		Foreground(colors.Primary).
		Bold(true).
		MarginBottom(1)
	
	s.ModalContent = lipgloss.NewStyle().
		Foreground(colors.Text)
	
	// Input styles
	s.Input = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colors.Border).
		Padding(0, 1)
	
	s.InputFocused = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colors.Primary).
		Padding(0, 1)
	
	s.InputLabel = lipgloss.NewStyle().
		Foreground(colors.TextMuted).
		MarginBottom(1)
	
	// Button styles
	s.Button = lipgloss.NewStyle().
		Foreground(colors.Text).
		Background(colors.BackgroundLight).
		Padding(0, 2).
		MarginRight(1)
	
	s.ButtonActive = lipgloss.NewStyle().
		Foreground(colors.TextBright).
		Background(colors.Primary).
		Bold(true).
		Padding(0, 2).
		MarginRight(1)
	
	s.ButtonDisabled = lipgloss.NewStyle().
		Foreground(colors.TextMuted).
		Background(colors.BackgroundDark).
		Padding(0, 2).
		MarginRight(1)
	
	// Syntax highlighting
	s.Keyword = lipgloss.NewStyle().
		Foreground(colors.Primary).
		Bold(true)
	
	s.String = lipgloss.NewStyle().
		Foreground(colors.Success)
	
	s.Number = lipgloss.NewStyle().
		Foreground(colors.Secondary)
	
	s.Comment = lipgloss.NewStyle().
		Foreground(colors.TextMuted).
		Italic(true)
	
	s.Function = lipgloss.NewStyle().
		Foreground(colors.Accent)
	
	// Messages
	s.ErrorText = lipgloss.NewStyle().
		Foreground(colors.Error)
	
	s.SuccessText = lipgloss.NewStyle().
		Foreground(colors.Success)
	
	s.InfoText = lipgloss.NewStyle().
		Foreground(colors.Info)
	
	s.WarningText = lipgloss.NewStyle().
		Foreground(colors.Warning)
	
	// Help
	s.HelpKey = lipgloss.NewStyle().
		Foreground(colors.Primary).
		Bold(true)
	
	s.HelpDesc = lipgloss.NewStyle().
		Foreground(colors.TextMuted)
	
	return s
}

// GetDraculaColors returns Dracula theme colors
func GetDraculaColors() ThemeColors {
	return ThemeColors{
		Primary:          lipgloss.Color("#bd93f9"),
		Secondary:        lipgloss.Color("#ff79c6"),
		Accent:           lipgloss.Color("#50fa7b"),
		Background:       lipgloss.Color("#282a36"),
		BackgroundDark:   lipgloss.Color("#21222c"),
		BackgroundLight:  lipgloss.Color("#44475a"),
		Text:             lipgloss.Color("#f8f8f2"),
		TextMuted:        lipgloss.Color("#6272a4"),
		TextBright:       lipgloss.Color("#ffffff"),
		Success:          lipgloss.Color("#50fa7b"),
		Warning:          lipgloss.Color("#ffb86c"),
		Error:            lipgloss.Color("#ff5555"),
		Info:             lipgloss.Color("#8be9fd"),
		Border:           lipgloss.Color("#44475a"),
		BorderFocus:      lipgloss.Color("#bd93f9"),
		Selection:        lipgloss.Color("#44475a"),
		GhostText:        lipgloss.Color("#6272a4"),
	}
}

// GetNordColors returns Nord theme colors
func GetNordColors() ThemeColors {
	return ThemeColors{
		Primary:          lipgloss.Color("#88c0d0"),
		Secondary:        lipgloss.Color("#81a1c1"),
		Accent:           lipgloss.Color("#a3be8c"),
		Background:       lipgloss.Color("#2e3440"),
		BackgroundDark:   lipgloss.Color("#242933"),
		BackgroundLight:  lipgloss.Color("#3b4252"),
		Text:             lipgloss.Color("#eceff4"),
		TextMuted:        lipgloss.Color("#4c566a"),
		TextBright:       lipgloss.Color("#ffffff"),
		Success:          lipgloss.Color("#a3be8c"),
		Warning:          lipgloss.Color("#ebcb8b"),
		Error:            lipgloss.Color("#bf616a"),
		Info:             lipgloss.Color("#88c0d0"),
		Border:           lipgloss.Color("#3b4252"),
		BorderFocus:      lipgloss.Color("#88c0d0"),
		Selection:        lipgloss.Color("#434c5e"),
		GhostText:        lipgloss.Color("#4c566a"),
	}
}

// GetDefaultColors returns default theme colors
func GetDefaultColors() ThemeColors {
	return ThemeColors{
		Primary:          lipgloss.Color("#007acc"),
		Secondary:        lipgloss.Color("#569cd6"),
		Accent:           lipgloss.Color("#4ec9b0"),
		Background:       lipgloss.Color("#1e1e1e"),
		BackgroundDark:   lipgloss.Color("#141414"),
		BackgroundLight:  lipgloss.Color("#2d2d2d"),
		Text:             lipgloss.Color("#d4d4d4"),
		TextMuted:        lipgloss.Color("#808080"),
		TextBright:       lipgloss.Color("#ffffff"),
		Success:          lipgloss.Color("#4ec9b0"),
		Warning:          lipgloss.Color("#dcdcaa"),
		Error:            lipgloss.Color("#f14c4c"),
		Info:             lipgloss.Color("#3794ff"),
		Border:           lipgloss.Color("#404040"),
		BorderFocus:      lipgloss.Color("#007acc"),
		Selection:        lipgloss.Color("#264f78"),
		GhostText:        lipgloss.Color("#5a5a5a"),
	}
}

// GetThemeColors returns colors for the specified theme
func GetThemeColors(themeName string) ThemeColors {
	switch themeName {
	case "dracula":
		return GetDraculaColors()
	case "nord":
		return GetNordColors()
	default:
		return GetDefaultColors()
	}
}
