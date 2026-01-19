package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Name     string `yaml:"name" mapstructure:"name"`
	Driver   string `yaml:"driver" mapstructure:"driver"` // postgres, mysql, sqlite
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
	User     string `yaml:"user" mapstructure:"user"`
	Password string `yaml:"password" mapstructure:"password"`
	Database string `yaml:"database" mapstructure:"database"`
	SSLMode  string `yaml:"sslmode" mapstructure:"sslmode"`
}

// AIConfig holds AI provider configuration
type AIConfig struct {
	Provider string `yaml:"provider" mapstructure:"provider"` // gemini, claude, openai, none
	APIKey   string `yaml:"api_key" mapstructure:"api_key"`
	Model    string `yaml:"model" mapstructure:"model"`
}

// Config is the main configuration structure
type Config struct {
	Theme           string           `yaml:"theme" mapstructure:"theme"`
	Editor          string           `yaml:"editor" mapstructure:"editor"` // External editor command
	AI              AIConfig         `yaml:"ai" mapstructure:"ai"`
	Connections     []DatabaseConfig `yaml:"connections" mapstructure:"connections"`
	ActiveConnIndex int              `yaml:"active_connection" mapstructure:"active_connection"`
	LastDatabase    string           `yaml:"last_database" mapstructure:"last_database"`
	LastTable       string           `yaml:"last_table" mapstructure:"last_table"`
	FirstRun        bool             `yaml:"first_run" mapstructure:"first_run"`
	KeyMap          KeyMap           `yaml:"keymap" mapstructure:"keymap"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Theme:  "dracula",
		Editor: "hx",
		AI: AIConfig{
			Provider: "none",
			Model:    "",
			APIKey:   "",
		},
		Connections:     []DatabaseConfig{},
		ActiveConnIndex: -1,
		FirstRun:        true,
		KeyMap:          DefaultKeyMap(),
	}
}

// GetConfigDir returns the configuration directory path
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "sqdesk"), nil
}

// GetConfigPath returns the full path to the config file
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.yaml"), nil
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(configDir, 0755)
}

// Load loads configuration from file
func Load() (*Config, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "config.yaml")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config for first run
		return DefaultConfig(), nil
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	// Set default editor if missing
	if cfg.Editor == "" {
		cfg.Editor = "hx"
	}

	return &cfg, nil
}

// Save saves configuration to file
func (c *Config) Save() error {
	if err := EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	viper.Set("theme", c.Theme)
	viper.Set("editor", c.Editor)
	viper.Set("ai", c.AI)
	viper.Set("connections", c.Connections)
	viper.Set("active_connection", c.ActiveConnIndex)
	viper.Set("last_database", c.LastDatabase)
	viper.Set("last_table", c.LastTable)
	viper.Set("first_run", c.FirstRun)
	viper.Set("keymap", c.KeyMap)

	return viper.WriteConfigAs(configPath)
}

// GetActiveConnection returns the currently active database connection config
func (c *Config) GetActiveConnection() *DatabaseConfig {
	if c.ActiveConnIndex < 0 || c.ActiveConnIndex >= len(c.Connections) {
		return nil
	}
	return &c.Connections[c.ActiveConnIndex]
}

// AddConnection adds a new database connection
func (c *Config) AddConnection(conn DatabaseConfig) {
	c.Connections = append(c.Connections, conn)
	if c.ActiveConnIndex < 0 {
		c.ActiveConnIndex = 0
	}
}

// RemoveConnection removes a database connection by index
func (c *Config) RemoveConnection(index int) error {
	if index < 0 || index >= len(c.Connections) {
		return fmt.Errorf("invalid connection index")
	}
	c.Connections = append(c.Connections[:index], c.Connections[index+1:]...)
	if c.ActiveConnIndex >= len(c.Connections) {
		c.ActiveConnIndex = len(c.Connections) - 1
	}
	return nil
}
