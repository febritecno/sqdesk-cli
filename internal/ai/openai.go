package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/febritecno/sqdesk/internal/db"
)

const openAIAPIURL = "https://api.openai.com/v1/chat/completions"

// OpenAIProvider implements Provider for OpenAI
type OpenAIProvider struct {
	apiKey string
	model  string
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAIProvider{
		apiKey: apiKey,
		model:  model,
	}
}

func (p *OpenAIProvider) GetProviderName() string {
	return "openai"
}

func (p *OpenAIProvider) GetModelName() string {
	return p.model
}

func (p *OpenAIProvider) IsConfigured() bool {
	return p.apiKey != ""
}

// OpenAIRequest represents the request body for OpenAI API
type OpenAIRequest struct {
	Model    string          `json:"model"`
	Messages []OpenAIMessage `json:"messages"`
}

// OpenAIMessage represents a message in OpenAI request
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents the response from OpenAI API
type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NL2SQL converts natural language to SQL using OpenAI
func (p *OpenAIProvider) NL2SQL(prompt string, schema *db.Schema) (string, error) {
	if !p.IsConfigured() {
		return "", fmt.Errorf("OpenAI API key not configured")
	}

	schemaContext := BuildSchemaContext(schema)
	
	systemPrompt := `You are a SQL expert. Convert the following natural language request into a valid SQL query.
Only respond with the SQL query, no explanations or markdown formatting.
Do not include any backticks or code blocks in your response.`

	userPrompt := schemaContext + "\n\nUser request: " + prompt

	return p.callAPI(systemPrompt, userPrompt)
}

// RefactorSQL modifies SQL based on instruction using OpenAI
func (p *OpenAIProvider) RefactorSQL(sql string, instruction string, schema *db.Schema) (string, error) {
	if !p.IsConfigured() {
		return "", fmt.Errorf("OpenAI API key not configured")
	}

	schemaContext := BuildSchemaContext(schema)
	
	systemPrompt := `You are a SQL expert. Modify the following SQL query based on the given instruction.
Only respond with the modified SQL query, no explanations or markdown formatting.
Do not include any backticks or code blocks in your response.`

	userPrompt := schemaContext + "\n\nOriginal SQL:\n" + sql + "\n\nInstruction: " + instruction

	return p.callAPI(systemPrompt, userPrompt)
}

func (p *OpenAIProvider) callAPI(systemPrompt, userPrompt string) (string, error) {
	reqBody := OpenAIRequest{
		Model: p.model,
		Messages: []OpenAIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", openAIAPIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

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

	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if openAIResp.Error != nil {
		return "", fmt.Errorf("API error: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	result := openAIResp.Choices[0].Message.Content
	// Clean up any markdown code blocks
	result = strings.TrimPrefix(result, "```sql")
	result = strings.TrimPrefix(result, "```")
	result = strings.TrimSuffix(result, "```")
	result = strings.TrimSpace(result)

	return result, nil
}
