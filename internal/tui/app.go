package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/febritecno/sqdesk/internal/config"
)

// App represents the SQDesk TUI application
type App struct {
	model   *Model
	program *tea.Program
}

// New creates a new SQDesk application
func New() (*App, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Create model
	model := NewModel(cfg)

	// Create Bubble Tea program
	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	return &App{
		model:   model,
		program: program,
	}, nil
}

// Run starts the TUI application
func (a *App) Run() error {
	// Try to connect to database if configured
	if !a.model.config.FirstRun {
		if err := a.model.Connect(); err != nil {
			a.model.statusMessage = "Connection failed: " + err.Error()
			a.model.isError = true
			// Continue anyway, user can reconfigure
		}
	}

	// Run the program
	if _, err := a.program.Run(); err != nil {
		return fmt.Errorf("application error: %w", err)
	}

	// Cleanup
	return a.model.Close()
}

// Close cleans up resources
func (a *App) Close() error {
	if a.model != nil {
		return a.model.Close()
	}
	return nil
}
