package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/febritecno/sqdesk-cli/internal/db"
)

const claudeAPIURL = "https://api.anthropic.com/v1/messages"

// ClaudeProvider implements Provider for Anthropic Claude
type ClaudeProvider struct {
	apiKey string
	model  string
}

// NewClaudeProvider creates a new Claude provider
func NewClaudeProvider(apiKey, model string) *ClaudeProvider {
	if model == "" {
		model = "claude-3-haiku-20240307"
	}
	return &ClaudeProvider{
		apiKey: apiKey,
		model:  model,
	}
}

func (p *ClaudeProvider) GetProviderName() string {
	return "claude"
}

func (p *ClaudeProvider) GetModelName() string {
	return p.model
}

func (p *ClaudeProvider) IsConfigured() bool {
	return p.apiKey != ""
}

// ClaudeRequest represents the request body for Claude API
type ClaudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []ClaudeMessage `json:"messages"`
}

// ClaudeMessage represents a message in Claude request
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeResponse represents the response from Claude API
type ClaudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NL2SQL converts natural language to SQL using Claude
func (p *ClaudeProvider) NL2SQL(prompt string, schema *db.Schema) (string, error) {
	if !p.IsConfigured() {
		return "", fmt.Errorf("Claude API key not configured")
	}

	schemaContext := BuildSchemaContext(schema)
	
	systemPrompt := `You are a SQL expert. Convert the following natural language request into a valid SQL query.
Only respond with the SQL query, no explanations or markdown formatting.
Do not include any backticks or code blocks in your response.

` + schemaContext + `

User request: ` + prompt

	return p.callAPI(systemPrompt)
}

// RefactorSQL modifies SQL based on instruction using Claude
func (p *ClaudeProvider) RefactorSQL(sql string, instruction string, schema *db.Schema) (string, error) {
	if !p.IsConfigured() {
		return "", fmt.Errorf("Claude API key not configured")
	}

	schemaContext := BuildSchemaContext(schema)
	
	systemPrompt := `You are a SQL expert. Modify the following SQL query based on the given instruction.
Only respond with the modified SQL query, no explanations or markdown formatting.
Do not include any backticks or code blocks in your response.

` + schemaContext + `

Original SQL:
` + sql + `

Instruction: ` + instruction

	return p.callAPI(systemPrompt)
}

func (p *ClaudeProvider) callAPI(prompt string) (string, error) {
	reqBody := ClaudeRequest{
		Model:     p.model,
		MaxTokens: 1024,
		Messages: []ClaudeMessage{
			{Role: "user", Content: prompt},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", claudeAPIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if claudeResp.Error != nil {
		return "", fmt.Errorf("API error: %s", claudeResp.Error.Message)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	result := claudeResp.Content[0].Text
	// Clean up any markdown code blocks
	result = strings.TrimPrefix(result, "```sql")
	result = strings.TrimPrefix(result, "```")
	result = strings.TrimSuffix(result, "```")
	result = strings.TrimSpace(result)

	return result, nil
}
