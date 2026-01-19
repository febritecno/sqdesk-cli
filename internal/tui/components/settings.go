package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SettingsTab represents a tab in settings
type SettingsTab int

const (
	SettingsTabTheme SettingsTab = iota
	SettingsTabAI
	SettingsTabConnections
)

// Settings component for settings modal
type Settings struct {
	visible     bool
	width       int
	height      int
	styles      SettingsStyles
	activeTab   SettingsTab
	themeIndex  int
	themes      []string
	
	// AI inputs
	aiProviderIndex int
	aiProviders     []string
	aiAPIKeyInput   textinput.Model
	aiModelInput    textinput.Model
	
	// Connection inputs
	connNameInput   textinput.Model
	connDriverIndex int
	connDrivers     []string
	connHostInput   textinput.Model
	connPortInput   textinput.Model
	connUserInput   textinput.Model
	connPassInput   textinput.Model
	connDBInput     textinput.Model
	
	focusedInput    int
}

// SettingsStyles holds styling for the settings
type SettingsStyles struct {
	Modal      lipgloss.Style
	Title      lipgloss.Style
	Tab        lipgloss.Style
	TabActive  lipgloss.Style
	Label      lipgloss.Style
	Input      lipgloss.Style
	InputFocus lipgloss.Style
	Button     lipgloss.Style
	ButtonActive lipgloss.Style
	Selected   lipgloss.Style
	Hint       lipgloss.Style
}

// NewSettings creates a new settings component
func NewSettings(styles SettingsStyles) Settings {
	// AI inputs
	aiKey := textinput.New()
	aiKey.Placeholder = "Enter API key..."
	aiKey.EchoMode = textinput.EchoPassword
	aiKey.Width = 40

	aiModel := textinput.New()
	aiModel.Placeholder = "e.g., gemini-1.5-flash"
	aiModel.Width = 40

	// Connection inputs
	connName := textinput.New()
	connName.Placeholder = "Connection name"
	connName.Width = 40

	connHost := textinput.New()
	connHost.Placeholder = "localhost"
	connHost.Width = 40

	connPort := textinput.New()
	connPort.Placeholder = "5432"
	connPort.Width = 10

	connUser := textinput.New()
	connUser.Placeholder = "username"
	connUser.Width = 40

	connPass := textinput.New()
	connPass.Placeholder = "password"
	connPass.EchoMode = textinput.EchoPassword
	connPass.Width = 40

	connDB := textinput.New()
	connDB.Placeholder = "database"
	connDB.Width = 40

	return Settings{
		visible:         false,
		styles:          styles,
		activeTab:       SettingsTabTheme,
		themes:          []string{"dracula", "nord", "default"},
		themeIndex:      0,
		aiProviders:     []string{"gemini", "claude", "openai", "none"},
		aiProviderIndex: 0,
		aiAPIKeyInput:   aiKey,
		aiModelInput:    aiModel,
		connNameInput:   connName,
		connDrivers:     []string{"postgres", "mysql", "sqlite"},
		connDriverIndex: 0,
		connHostInput:   connHost,
		connPortInput:   connPort,
		connUserInput:   connUser,
		connPassInput:   connPass,
		connDBInput:     connDB,
		focusedInput:    0,
	}
}

// Show shows the settings modal
func (s *Settings) Show() {
	s.visible = true
}

// Hide hides the settings modal
func (s *Settings) Hide() {
	s.visible = false
}

// IsVisible returns if settings is visible
func (s Settings) IsVisible() bool {
	return s.visible
}

// SetSize sets the settings dimensions
func (s *Settings) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// SetTheme sets the current theme index
func (s *Settings) SetTheme(theme string) {
	for i, t := range s.themes {
		if t == theme {
			s.themeIndex = i
			return
		}
	}
}

// GetSelectedTheme returns the selected theme name
func (s Settings) GetSelectedTheme() string {
	return s.themes[s.themeIndex]
}

// SetAIProvider sets the AI provider
func (s *Settings) SetAIProvider(provider string) {
	for i, p := range s.aiProviders {
		if p == provider {
			s.aiProviderIndex = i
			return
		}
	}
}

// GetSelectedProvider returns the selected AI provider
func (s Settings) GetSelectedProvider() string {
	return s.aiProviders[s.aiProviderIndex]
}

// SetAPIKey sets the API key
func (s *Settings) SetAPIKey(key string) {
	s.aiAPIKeyInput.SetValue(key)
}

// GetAPIKey returns the API key
func (s Settings) GetAPIKey() string {
	return s.aiAPIKeyInput.Value()
}

// SetModel sets the AI model
func (s *Settings) SetModel(model string) {
	s.aiModelInput.SetValue(model)
}

// GetModel returns the AI model
func (s Settings) GetModel() string {
	return s.aiModelInput.Value()
}

// GetConnectionConfig returns the connection configuration
func (s Settings) GetConnectionConfig() (name, driver, host, port, user, pass, db string) {
	return s.connNameInput.Value(),
		s.connDrivers[s.connDriverIndex],
		s.connHostInput.Value(),
		s.connPortInput.Value(),
		s.connUserInput.Value(),
		s.connPassInput.Value(),
		s.connDBInput.Value()
}

// NextTab moves to the next tab
func (s *Settings) NextTab() {
	s.activeTab = (s.activeTab + 1) % 3
	s.focusedInput = 0
}

// PrevTab moves to the previous tab
func (s *Settings) PrevTab() {
	if s.activeTab == 0 {
		s.activeTab = 2
	} else {
		s.activeTab--
	}
	s.focusedInput = 0
}

// Update handles input for settings
func (s Settings) Update(msg tea.Msg) (Settings, tea.Cmd) {
	if !s.visible {
		return s, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		// Check if any input is currently focused
		inputFocused := s.isInputFocused()

		// If input is focused, only intercept tab and arrow keys for navigation
		// Let other keys pass through to the input
		if inputFocused {
			switch key {
			case "tab":
				s.NextTab()
				return s, nil
			case "shift+tab":
				s.PrevTab()
				return s, nil
			case "up":
				s.navigateUp()
				return s, nil
			case "down":
				s.navigateDown()
				return s, nil
			default:
				// Pass all other keys to the focused input
				return s.updateFocusedInput(msg)
			}
		}

		// No input focused - handle navigation keys
		switch key {
		case "tab":
			s.NextTab()
			return s, nil
		case "shift+tab":
			s.PrevTab()
			return s, nil
		case "up":
			s.navigateUp()
			return s, nil
		case "down":
			s.navigateDown()
			return s, nil
		case "left":
			s.navigateLeft()
			return s, nil
		case "right":
			s.navigateRight()
			return s, nil
		}
	}

	return s, nil
}

// isInputFocused returns true if a text input is currently focused
func (s Settings) isInputFocused() bool {
	switch s.activeTab {
	case SettingsTabAI:
		return s.focusedInput == 1 || s.focusedInput == 2
	case SettingsTabConnections:
		return s.focusedInput == 0 || s.focusedInput >= 2
	}
	return false
}

// updateFocusedInput sends the message to the focused input
func (s Settings) updateFocusedInput(msg tea.Msg) (Settings, tea.Cmd) {
	var cmd tea.Cmd
	switch s.activeTab {
	case SettingsTabAI:
		if s.focusedInput == 1 {
			s.aiAPIKeyInput, cmd = s.aiAPIKeyInput.Update(msg)
		} else if s.focusedInput == 2 {
			s.aiModelInput, cmd = s.aiModelInput.Update(msg)
		}
	case SettingsTabConnections:
		switch s.focusedInput {
		case 0:
			s.connNameInput, cmd = s.connNameInput.Update(msg)
		case 2:
			s.connHostInput, cmd = s.connHostInput.Update(msg)
		case 3:
			s.connPortInput, cmd = s.connPortInput.Update(msg)
		case 4:
			s.connUserInput, cmd = s.connUserInput.Update(msg)
		case 5:
			s.connPassInput, cmd = s.connPassInput.Update(msg)
		case 6:
			s.connDBInput, cmd = s.connDBInput.Update(msg)
		}
	}
	return s, cmd
}

func (s *Settings) navigateUp() {
	if s.focusedInput > 0 {
		s.focusedInput--
	}
	s.updateInputFocus()
}

func (s *Settings) navigateDown() {
	maxInput := 0
	switch s.activeTab {
	case SettingsTabTheme:
		maxInput = 0
	case SettingsTabAI:
		maxInput = 2
	case SettingsTabConnections:
		maxInput = 6
	}
	if s.focusedInput < maxInput {
		s.focusedInput++
	}
	s.updateInputFocus()
}

func (s *Settings) navigateLeft() {
	switch s.activeTab {
	case SettingsTabTheme:
		if s.themeIndex > 0 {
			s.themeIndex--
		}
	case SettingsTabAI:
		if s.focusedInput == 0 && s.aiProviderIndex > 0 {
			s.aiProviderIndex--
		}
	case SettingsTabConnections:
		if s.focusedInput == 1 && s.connDriverIndex > 0 {
			s.connDriverIndex--
		}
	}
}

func (s *Settings) navigateRight() {
	switch s.activeTab {
	case SettingsTabTheme:
		if s.themeIndex < len(s.themes)-1 {
			s.themeIndex++
		}
	case SettingsTabAI:
		if s.focusedInput == 0 && s.aiProviderIndex < len(s.aiProviders)-1 {
			s.aiProviderIndex++
		}
	case SettingsTabConnections:
		if s.focusedInput == 1 && s.connDriverIndex < len(s.connDrivers)-1 {
			s.connDriverIndex++
		}
	}
}

func (s *Settings) updateInputFocus() {
	// Blur all inputs
	s.aiAPIKeyInput.Blur()
	s.aiModelInput.Blur()
	s.connNameInput.Blur()
	s.connHostInput.Blur()
	s.connPortInput.Blur()
	s.connUserInput.Blur()
	s.connPassInput.Blur()
	s.connDBInput.Blur()

	// Focus the current input
	switch s.activeTab {
	case SettingsTabAI:
		if s.focusedInput == 1 {
			s.aiAPIKeyInput.Focus()
		} else if s.focusedInput == 2 {
			s.aiModelInput.Focus()
		}
	case SettingsTabConnections:
		switch s.focusedInput {
		case 0:
			s.connNameInput.Focus()
		case 2:
			s.connHostInput.Focus()
		case 3:
			s.connPortInput.Focus()
		case 4:
			s.connUserInput.Focus()
		case 5:
			s.connPassInput.Focus()
		case 6:
			s.connDBInput.Focus()
		}
	}
}

// View renders the settings modal
func (s Settings) View() string {
	if !s.visible {
		return ""
	}

	// Ensure minimum width
	width := s.width
	if width < 60 {
		width = 60
	}

	content := s.styles.Title.Render("⚙️  Settings") + "\n\n"

	// Tabs
	tabs := ""
	tabNames := []string{"Theme", "AI", "Connections"}
	for i, name := range tabNames {
		style := s.styles.Tab
		if SettingsTab(i) == s.activeTab {
			style = s.styles.TabActive
		}
		tabs += style.Render(name) + " "
	}
	content += tabs + "\n\n"

	// Tab content
	switch s.activeTab {
	case SettingsTabTheme:
		content += s.viewThemeTab()
	case SettingsTabAI:
		content += s.viewAITab()
	case SettingsTabConnections:
		content += s.viewConnectionsTab()
	}

	content += "\n\n" + s.styles.Hint.Render("Tab: switch tabs • ↑↓: navigate • ←→: select • Enter: save • Esc: close")

	return s.styles.Modal.
		Width(width).
		Padding(1, 2).
		Render(content)
}

func (s Settings) viewThemeTab() string {
	content := s.styles.Label.Render("Select Theme:") + "\n\n"

	for i, theme := range s.themes {
		style := s.styles.Button
		if i == s.themeIndex {
			style = s.styles.Selected
		}
		content += style.Render(theme) + "  "
	}

	return content
}

func (s Settings) viewAITab() string {
	content := ""

	// Provider selection
	label := s.styles.Label
	if s.focusedInput == 0 {
		label = s.styles.Selected
	}
	content += label.Render("Provider:") + "\n"
	for i, provider := range s.aiProviders {
		style := s.styles.Button
		if i == s.aiProviderIndex {
			style = s.styles.ButtonActive
		}
		content += style.Render(provider) + " "
	}
	content += "\n\n"

	// API Key
	label = s.styles.Label
	if s.focusedInput == 1 {
		label = s.styles.Selected
	}
	content += label.Render("API Key:") + "\n"
	content += s.aiAPIKeyInput.View() + "\n\n"

	// Model
	label = s.styles.Label
	if s.focusedInput == 2 {
		label = s.styles.Selected
	}
	content += label.Render("Model:") + "\n"
	content += s.aiModelInput.View()

	return content
}

func (s Settings) viewConnectionsTab() string {
	content := ""

	fields := []struct {
		idx   int
		label string
		input string
	}{
		{0, "Connection Name:", s.connNameInput.View()},
	}

	for _, f := range fields {
		label := s.styles.Label
		if s.focusedInput == f.idx {
			label = s.styles.Selected
		}
		content += label.Render(f.label) + "\n" + f.input + "\n\n"
	}

	// Driver selection
	label := s.styles.Label
	if s.focusedInput == 1 {
		label = s.styles.Selected
	}
	content += label.Render("Driver:") + "\n"
	for i, driver := range s.connDrivers {
		style := s.styles.Button
		if i == s.connDriverIndex {
			style = s.styles.ButtonActive
		}
		content += style.Render(driver) + " "
	}
	content += "\n\n"

	// Other fields
	moreFields := []struct {
		idx   int
		label string
		input string
	}{
		{2, "Host:", s.connHostInput.View()},
		{3, "Port:", s.connPortInput.View()},
		{4, "User:", s.connUserInput.View()},
		{5, "Password:", s.connPassInput.View()},
		{6, "Database:", s.connDBInput.View()},
	}

	for _, f := range moreFields {
		label := s.styles.Label
		if s.focusedInput == f.idx {
			label = s.styles.Selected
		}
		content += label.Render(f.label) + "\n" + f.input + "\n\n"
	}

	return content
}
