package components

import (
	"fmt"
	"strings"
	
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AIPrompt component for AI input modal
type AIPrompt struct {
	input          textinput.Model
	visible        bool
	width          int
	height         int
	styles         AIPromptStyles
	mode           AIPromptMode
	onSubmit       func(string)
	selectedText   string
	hasContext     bool
}

// AIPromptMode indicates the type of AI operation
type AIPromptMode int

const (
	AIPromptModeNL2SQL AIPromptMode = iota
	AIPromptModeRefactor
)

// AIPromptStyles holds styling for the AI prompt
type AIPromptStyles struct {
	Modal       lipgloss.Style
	Title       lipgloss.Style
	Input       lipgloss.Style
	Hint        lipgloss.Style
	ModeNL2SQL  lipgloss.Style
	ModeRefactor lipgloss.Style
}

// NewAIPrompt creates a new AI prompt component
func NewAIPrompt(styles AIPromptStyles) AIPrompt {
	ti := textinput.New()
	ti.Placeholder = "Describe what you want..."
	ti.CharLimit = 500
	ti.Width = 50

	return AIPrompt{
		input:   ti,
		visible: false,
		styles:  styles,
		mode:    AIPromptModeNL2SQL,
	}
}

// Show shows the AI prompt modal
func (a *AIPrompt) Show(mode AIPromptMode) {
	a.visible = true
	a.mode = mode
	a.input.SetValue("")
	a.input.Focus()
	
	if mode == AIPromptModeNL2SQL {
		a.input.Placeholder = "Describe the query you want to generate..."
	} else {
		a.input.Placeholder = "Describe how to modify the SQL..."
	}
}

// SetContext sets the selected text context
func (a *AIPrompt) SetContext(text string) {
	a.selectedText = text
	a.hasContext = len(text) > 0
}

// ClearContext clears the context
func (a *AIPrompt) ClearContext() {
	a.selectedText = ""
	a.hasContext = false
}

// Hide hides the AI prompt modal
func (a *AIPrompt) Hide() {
	a.visible = false
	a.input.Blur()
}

// IsVisible returns if the prompt is visible
func (a AIPrompt) IsVisible() bool {
	return a.visible
}

// GetMode returns the current mode
func (a AIPrompt) GetMode() AIPromptMode {
	return a.mode
}

// GetValue returns the input value
func (a AIPrompt) GetValue() string {
	return a.input.Value()
}

// SetSize sets the prompt dimensions
func (a *AIPrompt) SetSize(width, height int) {
	a.width = width
	a.height = height
	a.input.Width = width - 10
}

// Update handles input for the AI prompt
func (a AIPrompt) Update(msg tea.Msg) (AIPrompt, tea.Cmd) {
	if !a.visible {
		return a, nil
	}

	var cmd tea.Cmd
	a.input, cmd = a.input.Update(msg)
	return a, cmd
}

// View renders the AI prompt
func (a AIPrompt) View() string {
	if !a.visible {
		return ""
	}

	// Title based on mode
	title := "🤖 AI: Generate SQL"
	if a.mode == AIPromptModeRefactor {
		title = "🔧 AI: Refactor SQL"
	}

	// Ensure minimum width
	width := a.width
	if width < 60 {
		width = 60
	}

	content := a.styles.Title.Render(title) + "\n\n"
	
	// Show context info if available
	if a.hasContext {
		lines := len(strings.Split(a.selectedText, "\n"))
		chars := len(a.selectedText)
		contextInfo := a.styles.Hint.Render(fmt.Sprintf("📝 Working with selection: %d lines, %d chars", lines, chars))
		content += contextInfo + "\n\n"
	}
	
	content += a.input.View() + "\n\n"
	content += a.styles.Hint.Render("Press Enter to submit • Esc to cancel")

	return a.styles.Modal.
		Width(width).
		Padding(1, 2).
		Render(content)
}

