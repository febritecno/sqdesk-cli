package ai

import (
	"github.com/febritecno/sqdesk/internal/db"
)

// Provider interface for AI operations
type Provider interface {
	// NL2SQL converts natural language to SQL query
	NL2SQL(prompt string, schema *db.Schema) (string, error)
	
	// RefactorSQL modifies SQL based on instruction
	RefactorSQL(sql string, instruction string, schema *db.Schema) (string, error)
	
	// GetProviderName returns the provider name
	GetProviderName() string
	
	// GetModelName returns the model name
	GetModelName() string
	
	// IsConfigured checks if the provider is properly configured
	IsConfigured() bool
}

// NewProvider creates a new AI provider based on provider name
func NewProvider(providerName, apiKey, model string) (Provider, error) {
	switch providerName {
	case "gemini":
		return NewGeminiProvider(apiKey, model), nil
	case "claude":
		return NewClaudeProvider(apiKey, model), nil
	case "openai":
		return NewOpenAIProvider(apiKey, model), nil
	default:
		return NewNoopProvider(), nil
	}
}

// NoopProvider is a no-operation provider when AI is disabled
type NoopProvider struct{}

// NewNoopProvider creates a no-op provider
func NewNoopProvider() *NoopProvider {
	return &NoopProvider{}
}

func (p *NoopProvider) NL2SQL(prompt string, schema *db.Schema) (string, error) {
	return "", nil
}

func (p *NoopProvider) RefactorSQL(sql string, instruction string, schema *db.Schema) (string, error) {
	return sql, nil
}

func (p *NoopProvider) GetProviderName() string {
	return "none"
}

func (p *NoopProvider) GetModelName() string {
	return ""
}

func (p *NoopProvider) IsConfigured() bool {
	return false
}

// BuildSchemaContext creates a text representation of the schema for AI prompts
func BuildSchemaContext(schema *db.Schema) string {
	if schema == nil {
		return ""
	}

	var result string
	result = "Database Schema:\n"
	
	for tableName, table := range schema.Tables {
		result += "\nTable: " + tableName + "\n"
		result += "Columns:\n"
		for _, col := range table.Columns {
			pkMarker := ""
			if col.IsPK {
				pkMarker = " (PRIMARY KEY)"
			}
			nullMarker := ""
			if col.Nullable {
				nullMarker = " NULL"
			} else {
				nullMarker = " NOT NULL"
			}
			result += "  - " + col.Name + " " + col.Type + nullMarker + pkMarker + "\n"
		}
	}
	
	return result
}
