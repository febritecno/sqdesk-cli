package config

import "github.com/charmbracelet/lipgloss"

// Theme holds color definitions for the TUI
type Theme struct {
	Name string

	// Primary colors
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent    lipgloss.Color

	// Background colors
	Background       lipgloss.Color
	BackgroundDark   lipgloss.Color
	BackgroundLight  lipgloss.Color

	// Text colors
	Text       lipgloss.Color
	TextMuted  lipgloss.Color
	TextBright lipgloss.Color

	// Semantic colors
	Success lipgloss.Color
	Warning lipgloss.Color
	Error   lipgloss.Color
	Info    lipgloss.Color

	// UI specific
	Border       lipgloss.Color
	BorderFocus  lipgloss.Color
	Selection    lipgloss.Color
	Cursor       lipgloss.Color
	GhostText    lipgloss.Color

	// Syntax highlighting
	Keyword  lipgloss.Color
	String   lipgloss.Color
	Number   lipgloss.Color
	Comment  lipgloss.Color
	Function lipgloss.Color
}

// Dracula theme
var Dracula = Theme{
	Name:             "dracula",
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
	Cursor:           lipgloss.Color("#f8f8f2"),
	GhostText:        lipgloss.Color("#6272a4"),
	Keyword:          lipgloss.Color("#ff79c6"),
	String:           lipgloss.Color("#f1fa8c"),
	Number:           lipgloss.Color("#bd93f9"),
	Comment:          lipgloss.Color("#6272a4"),
	Function:         lipgloss.Color("#50fa7b"),
}

// Nord theme
var Nord = Theme{
	Name:             "nord",
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
	Cursor:           lipgloss.Color("#eceff4"),
	GhostText:        lipgloss.Color("#4c566a"),
	Keyword:          lipgloss.Color("#81a1c1"),
	String:           lipgloss.Color("#a3be8c"),
	Number:           lipgloss.Color("#b48ead"),
	Comment:          lipgloss.Color("#616e88"),
	Function:         lipgloss.Color("#88c0d0"),
}

// SystemDefault theme (neutral colors)
var SystemDefault = Theme{
	Name:             "default",
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
	Cursor:           lipgloss.Color("#aeafad"),
	GhostText:        lipgloss.Color("#5a5a5a"),
	Keyword:          lipgloss.Color("#569cd6"),
	String:           lipgloss.Color("#ce9178"),
	Number:           lipgloss.Color("#b5cea8"),
	Comment:          lipgloss.Color("#6a9955"),
	Function:         lipgloss.Color("#dcdcaa"),
}

// AvailableThemes returns all available theme names
var AvailableThemes = []string{"dracula", "nord", "default"}

// GetTheme returns the theme by name
func GetTheme(name string) Theme {
	switch name {
	case "dracula":
		return Dracula
	case "nord":
		return Nord
	case "default":
		return SystemDefault
	default:
		return Dracula
	}
}
