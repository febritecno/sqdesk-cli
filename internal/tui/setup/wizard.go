package setup

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/febritecno/sqdesk-cli/internal/config"
)

// WizardStep represents the current step in the wizard
type WizardStep int

const (
	StepWelcome WizardStep = iota
	StepTheme
	StepAI
	StepConnection
	StepComplete
)

// Wizard handles the first-run setup experience
type Wizard struct {
	step         WizardStep
	width        int
	height       int
	styles       WizardStyles
	
	// Theme selection
	themes       []string
	themeIndex   int
	
	// AI configuration
	aiProviders  []string
	aiIndex      int
	aiKeyInput   textinput.Model
	aiModelInput textinput.Model
	
	// Connection configuration
	connInputs   []textinput.Model
	connLabels   []string
	connDrivers  []string
	connDriverIdx int
	focusedInput int
	
	// Result
	config       *config.Config
}

// WizardStyles holds styling for the wizard
type WizardStyles struct {
	Container  lipgloss.Style
	Title      lipgloss.Style
	Subtitle   lipgloss.Style
	ASCII      lipgloss.Style
	Text       lipgloss.Style
	Selected   lipgloss.Style
	Unselected lipgloss.Style
	Input      lipgloss.Style
	InputFocus lipgloss.Style
	Label      lipgloss.Style
	Button     lipgloss.Style
	ButtonActive lipgloss.Style
	Hint       lipgloss.Style
	Success    lipgloss.Style
}

// ASCII art for welcome screen
const asciiLogo = `
   _____ ____  ____            _    
  / ____|  _ \|  _ \          | |   
 | (___ | |_) | | | | ___  ___| | __
  \___ \|  _ <| | | |/ _ \/ __| |/ /
  ____) | |_) | |_| |  __/\__ \   < 
 |_____/|____/|____/ \___||___/_|\_\
                                    
`

// NewWizard creates a new setup wizard
func NewWizard(styles WizardStyles) *Wizard {
	// AI key input
	aiKey := textinput.New()
	aiKey.Placeholder = "Enter your API key..."
	aiKey.EchoMode = textinput.EchoPassword
	aiKey.Width = 50

	aiModel := textinput.New()
	aiModel.Placeholder = "e.g., gemini-1.5-flash"
	aiModel.Width = 50

	// Connection inputs
	connInputs := make([]textinput.Model, 6)
	connLabels := []string{"Connection Name", "Host", "Port", "User", "Password", "Database"}
	placeholders := []string{"My Database", "localhost", "5432", "postgres", "password", "mydb"}

	for i := range connInputs {
		connInputs[i] = textinput.New()
		connInputs[i].Placeholder = placeholders[i]
		connInputs[i].Width = 50
		if connLabels[i] == "Password" {
			connInputs[i].EchoMode = textinput.EchoPassword
		}
	}
	connInputs[0].Focus()

	return &Wizard{
		step:          StepWelcome,
		styles:        styles,
		themes:        []string{"Dracula", "Nord", "Default"},
		themeIndex:    0,
		aiProviders:   []string{"Gemini", "Claude", "OpenAI", "Skip"},
		aiIndex:       0,
		aiKeyInput:    aiKey,
		aiModelInput:  aiModel,
		connInputs:    connInputs,
		connLabels:    connLabels,
		connDrivers:   []string{"PostgreSQL", "MySQL", "SQLite"},
		connDriverIdx: 0,
		focusedInput:  0,
		config:        config.DefaultConfig(),
	}
}

// SetSize sets the wizard dimensions
func (w *Wizard) SetSize(width, height int) {
	w.width = width
	w.height = height
}

// IsComplete returns if the wizard is complete
func (w *Wizard) IsComplete() bool {
	return w.step == StepComplete
}

// GetConfig returns the configured settings
func (w *Wizard) GetConfig() *config.Config {
	return w.config
}

// Update handles input for the wizard
func (w *Wizard) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return w.nextStep()
		case "esc":
			if w.step > StepWelcome {
				w.step--
			}
			return nil
		case "up", "k":
			w.navigateUp()
		case "down", "j":
			w.navigateDown()
		case "left", "h":
			w.navigateLeft()
		case "right", "l":
			w.navigateRight()
		case "tab":
			w.navigateDown()
		case "shift+tab":
			w.navigateUp()
		default:
			return w.updateInputs(msg)
		}
	}
	return nil
}

func (w *Wizard) nextStep() tea.Cmd {
	switch w.step {
	case StepWelcome:
		w.step = StepTheme
	case StepTheme:
		// Save theme
		w.config.Theme = strings.ToLower(w.themes[w.themeIndex])
		w.step = StepAI
	case StepAI:
		// Save AI config
		if w.aiIndex < 3 { // Not "Skip"
			w.config.AI.Provider = strings.ToLower(w.aiProviders[w.aiIndex])
			w.config.AI.APIKey = w.aiKeyInput.Value()
			w.config.AI.Model = w.aiModelInput.Value()
		} else {
			w.config.AI.Provider = "none"
		}
		w.step = StepConnection
	case StepConnection:
		// Save connection config
		port, _ := strconv.Atoi(w.connInputs[2].Value())
		if port == 0 {
			port = 5432
		}
		
		driver := strings.ToLower(w.connDrivers[w.connDriverIdx])
		if driver == "postgresql" {
			driver = "postgres"
		}
		
		conn := config.DatabaseConfig{
			Name:     w.connInputs[0].Value(),
			Driver:   driver,
			Host:     w.connInputs[1].Value(),
			Port:     port,
			User:     w.connInputs[3].Value(),
			Password: w.connInputs[4].Value(),
			Database: w.connInputs[5].Value(),
		}
		
		if conn.Name == "" {
			conn.Name = "Default"
		}
		if conn.Host == "" {
			conn.Host = "localhost"
		}
		
		w.config.AddConnection(conn)
		w.config.FirstRun = false
		w.step = StepComplete
	}
	return nil
}

func (w *Wizard) navigateUp() {
	switch w.step {
	case StepTheme:
		// Themes are horizontal
	case StepAI:
		if w.focusedInput > 0 {
			w.focusedInput--
			w.updateFocus()
		}
	case StepConnection:
		if w.focusedInput > 0 {
			w.focusedInput--
			w.updateConnFocus()
		}
	}
}

func (w *Wizard) navigateDown() {
	switch w.step {
	case StepTheme:
		// Themes are horizontal
	case StepAI:
		if w.focusedInput < 2 {
			w.focusedInput++
			w.updateFocus()
		}
	case StepConnection:
		if w.focusedInput < 6 {
			w.focusedInput++
			w.updateConnFocus()
		}
	}
}

func (w *Wizard) navigateLeft() {
	switch w.step {
	case StepTheme:
		if w.themeIndex > 0 {
			w.themeIndex--
		}
	case StepAI:
		if w.focusedInput == 0 && w.aiIndex > 0 {
			w.aiIndex--
		}
	case StepConnection:
		if w.focusedInput == 0 && w.connDriverIdx > 0 {
			w.connDriverIdx--
		}
	}
}

func (w *Wizard) navigateRight() {
	switch w.step {
	case StepTheme:
		if w.themeIndex < len(w.themes)-1 {
			w.themeIndex++
		}
	case StepAI:
		if w.focusedInput == 0 && w.aiIndex < len(w.aiProviders)-1 {
			w.aiIndex++
		}
	case StepConnection:
		if w.focusedInput == 0 && w.connDriverIdx < len(w.connDrivers)-1 {
			w.connDriverIdx++
		}
	}
}

func (w *Wizard) updateFocus() {
	w.aiKeyInput.Blur()
	w.aiModelInput.Blur()
	
	if w.focusedInput == 1 {
		w.aiKeyInput.Focus()
	} else if w.focusedInput == 2 {
		w.aiModelInput.Focus()
	}
}

func (w *Wizard) updateConnFocus() {
	for i := range w.connInputs {
		w.connInputs[i].Blur()
	}
	
	if w.focusedInput > 0 && w.focusedInput <= len(w.connInputs) {
		w.connInputs[w.focusedInput-1].Focus()
	}
}

func (w *Wizard) updateInputs(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	
	switch w.step {
	case StepAI:
		if w.focusedInput == 1 {
			w.aiKeyInput, cmd = w.aiKeyInput.Update(msg)
		} else if w.focusedInput == 2 {
			w.aiModelInput, cmd = w.aiModelInput.Update(msg)
		}
	case StepConnection:
		if w.focusedInput > 0 && w.focusedInput <= len(w.connInputs) {
			w.connInputs[w.focusedInput-1], cmd = w.connInputs[w.focusedInput-1].Update(msg)
		}
	}
	
	return cmd
}

// View renders the wizard
func (w *Wizard) View() string {
	var content string

	switch w.step {
	case StepWelcome:
		content = w.viewWelcome()
	case StepTheme:
		content = w.viewTheme()
	case StepAI:
		content = w.viewAI()
	case StepConnection:
		content = w.viewConnection()
	case StepComplete:
		content = w.viewComplete()
	}

	return w.styles.Container.
		Width(w.width).
		Height(w.height).
		Render(content)
}

func (w *Wizard) viewWelcome() string {
	var b strings.Builder

	b.WriteString(w.styles.ASCII.Render(asciiLogo))
	b.WriteString("\n")
	b.WriteString(w.styles.Title.Render("Welcome to SQDesk!"))
	b.WriteString("\n\n")
	b.WriteString(w.styles.Text.Render("A lightweight, intelligent database client for your terminal."))
	b.WriteString("\n")
	b.WriteString(w.styles.Text.Render("Let's set things up quickly."))
	b.WriteString("\n\n")
	b.WriteString(w.styles.Hint.Render("Press Enter to continue..."))

	return b.String()
}

func (w *Wizard) viewTheme() string {
	var b strings.Builder

	b.WriteString(w.styles.Title.Render("🎨 Choose Your Style"))
	b.WriteString("\n\n")
	b.WriteString(w.styles.Text.Render("Select a color theme for the interface:"))
	b.WriteString("\n\n")

	for i, theme := range w.themes {
		style := w.styles.Unselected
		prefix := "  "
		if i == w.themeIndex {
			style = w.styles.Selected
			prefix = "▸ "
		}
		b.WriteString(style.Render(prefix + theme))
		b.WriteString("  ")
	}

	b.WriteString("\n\n")
	b.WriteString(w.styles.Hint.Render("← → to select • Enter to continue • Esc to go back"))

	return b.String()
}

func (w *Wizard) viewAI() string {
	var b strings.Builder

	b.WriteString(w.styles.Title.Render("🤖 AI Configuration"))
	b.WriteString("\n\n")
	b.WriteString(w.styles.Text.Render("Configure AI for intelligent SQL generation:"))
	b.WriteString("\n\n")

	// Provider selection
	label := w.styles.Label
	if w.focusedInput == 0 {
		label = w.styles.Selected
	}
	b.WriteString(label.Render("Provider:"))
	b.WriteString("\n")

	for i, provider := range w.aiProviders {
		style := w.styles.Unselected
		if i == w.aiIndex {
			style = w.styles.Selected
		}
		b.WriteString(style.Render(provider))
		b.WriteString("  ")
	}
	b.WriteString("\n\n")

	// Only show API fields if not skipping
	if w.aiIndex < 3 {
		// API Key
		label = w.styles.Label
		if w.focusedInput == 1 {
			label = w.styles.Selected
		}
		b.WriteString(label.Render("API Key:"))
		b.WriteString("\n")
		b.WriteString(w.aiKeyInput.View())
		b.WriteString("\n\n")

		// Model
		label = w.styles.Label
		if w.focusedInput == 2 {
			label = w.styles.Selected
		}
		b.WriteString(label.Render("Model (optional):"))
		b.WriteString("\n")
		b.WriteString(w.aiModelInput.View())
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(w.styles.Hint.Render("← → to select provider • ↑↓ to navigate • Enter to continue"))

	return b.String()
}

func (w *Wizard) viewConnection() string {
	var b strings.Builder

	b.WriteString(w.styles.Title.Render("🗄️  Database Connection"))
	b.WriteString("\n\n")
	b.WriteString(w.styles.Text.Render("Set up your first database connection:"))
	b.WriteString("\n\n")

	// Driver selection
	label := w.styles.Label
	if w.focusedInput == 0 {
		label = w.styles.Selected
	}
	b.WriteString(label.Render("Database Type:"))
	b.WriteString("\n")

	for i, driver := range w.connDrivers {
		style := w.styles.Unselected
		if i == w.connDriverIdx {
			style = w.styles.Selected
		}
		b.WriteString(style.Render(driver))
		b.WriteString("  ")
	}
	b.WriteString("\n\n")

	// Connection fields
	for i, input := range w.connInputs {
		label := w.styles.Label
		if w.focusedInput == i+1 {
			label = w.styles.Selected
		}
		b.WriteString(label.Render(w.connLabels[i] + ":"))
		b.WriteString("\n")
		b.WriteString(input.View())
		b.WriteString("\n\n")
	}

	b.WriteString(w.styles.Hint.Render("↑↓ to navigate • ← → to select driver • Enter to finish"))

	return b.String()
}

func (w *Wizard) viewComplete() string {
	var b strings.Builder

	b.WriteString(w.styles.Success.Render("✓ Setup Complete!"))
	b.WriteString("\n\n")
	b.WriteString(w.styles.Text.Render("SQDesk is ready to use."))
	b.WriteString("\n\n")
	b.WriteString(w.styles.Text.Render("Quick tips:"))
	b.WriteString("\n")
	b.WriteString(w.styles.Text.Render("• Press Ctrl+Enter to run queries"))
	b.WriteString("\n")
	b.WriteString(w.styles.Text.Render("• Press Ctrl+G for AI-powered SQL generation"))
	b.WriteString("\n")
	b.WriteString(w.styles.Text.Render("• Press F2 to open settings anytime"))
	b.WriteString("\n\n")
	b.WriteString(w.styles.Hint.Render("Press Enter to start..."))

	return b.String()
}
