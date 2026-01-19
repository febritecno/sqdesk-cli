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

const geminiAPIURL = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s"

// GeminiProvider implements Provider for Google Gemini
type GeminiProvider struct {
	apiKey string
	model  string
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(apiKey, model string) *GeminiProvider {
	if model == "" {
		model = "gemini-1.5-flash"
	}
	return &GeminiProvider{
		apiKey: apiKey,
		model:  model,
	}
}

func (p *GeminiProvider) GetProviderName() string {
	return "gemini"
}

func (p *GeminiProvider) GetModelName() string {
	return p.model
}

func (p *GeminiProvider) IsConfigured() bool {
	return p.apiKey != ""
}

// GeminiRequest represents the request body for Gemini API
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

// GeminiContent represents content in Gemini request
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart represents a part in Gemini content
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiResponse represents the response from Gemini API
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NL2SQL converts natural language to SQL using Gemini
func (p *GeminiProvider) NL2SQL(prompt string, schema *db.Schema) (string, error) {
	if !p.IsConfigured() {
		return "", fmt.Errorf("Gemini API key not configured")
	}

	schemaContext := BuildSchemaContext(schema)
	
	systemPrompt := `You are a SQL expert. Convert the following natural language request into a valid SQL query.
Only respond with the SQL query, no explanations or markdown formatting.
Do not include any backticks or code blocks in your response.

` + schemaContext + `

User request: ` + prompt

	return p.callAPI(systemPrompt)
}

// RefactorSQL modifies SQL based on instruction using Gemini
func (p *GeminiProvider) RefactorSQL(sql string, instruction string, schema *db.Schema) (string, error) {
	if !p.IsConfigured() {
		return "", fmt.Errorf("Gemini API key not configured")
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

func (p *GeminiProvider) callAPI(prompt string) (string, error) {
	url := fmt.Sprintf(geminiAPIURL, p.model, p.apiKey)

	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if geminiResp.Error != nil {
		return "", fmt.Errorf("API error: %s", geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	result := geminiResp.Candidates[0].Content.Parts[0].Text
	// Clean up any markdown code blocks
	result = strings.TrimPrefix(result, "```sql")
	result = strings.TrimPrefix(result, "```")
	result = strings.TrimSuffix(result, "```")
	result = strings.TrimSpace(result)

	return result, nil
}
