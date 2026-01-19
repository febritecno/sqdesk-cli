package sources

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/febritecno/sqdesk-cli/internal/ai"
	"github.com/febritecno/sqdesk-cli/internal/completion"
)

// AISource provides AI-powered completions
type AISource struct {
	provider     ai.Provider
	cache        map[string]cacheEntry
	cacheMu      sync.RWMutex
	cacheTimeout time.Duration
	enabled      bool
}

type cacheEntry struct {
	items     []completion.CompletionItem
	timestamp time.Time
}

// NewAISource creates a new AI source
func NewAISource() *AISource {
	return &AISource{
		cache:        make(map[string]cacheEntry),
		cacheTimeout: 5 * time.Minute,
		enabled:      false,
	}
}

// Name returns the source name
func (s *AISource) Name() string {
	return "ai"
}

// Priority returns the source priority
func (s *AISource) Priority() int {
	return 30 // Lower priority - AI is slower
}

// SetProvider sets the AI provider
func (s *AISource) SetProvider(provider ai.Provider) {
	s.provider = provider
	s.enabled = provider != nil
}

// SetEnabled enables or disables AI completions
func (s *AISource) SetEnabled(enabled bool) {
	s.enabled = enabled
}

// Complete returns AI-powered completions
func (s *AISource) Complete(ctx completion.Context) ([]completion.CompletionItem, error) {
	if !s.enabled || s.provider == nil {
		return nil, nil
	}
	
	// Skip if word is too short
	if len(ctx.Word) < 3 {
		return nil, nil
	}
	
	// Check cache
	cacheKey := s.getCacheKey(ctx)
	if items := s.getFromCache(cacheKey); items != nil {
		return items, nil
	}
	
	// Build prompt for AI
	prompt := s.buildPrompt(ctx)
	
	// Call AI provider with timeout - use NL2SQL as generic completion
	aiCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	// Use a channel to handle timeout
	type result struct {
		response string
		err      error
	}
	resultChan := make(chan result, 1)
	
	go func() {
		response, err := s.provider.NL2SQL(prompt, nil)
		resultChan <- result{response, err}
	}()
	
	select {
	case <-aiCtx.Done():
		return nil, aiCtx.Err()
	case res := <-resultChan:
		if res.err != nil {
			return nil, res.err
		}
		// Parse response into completion items
		items := s.parseResponse(res.response, ctx)
		
		// Cache results
		s.cacheResult(cacheKey, items)
		
		return items, nil
	}
}

// getCacheKey generates a cache key for the context
func (s *AISource) getCacheKey(ctx completion.Context) string {
	// Use line prefix + word as cache key
	return ctx.LinePrefix + "|" + ctx.Word
}

// getFromCache retrieves cached items if valid
func (s *AISource) getFromCache(key string) []completion.CompletionItem {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	
	entry, ok := s.cache[key]
	if !ok {
		return nil
	}
	
	if time.Since(entry.timestamp) > s.cacheTimeout {
		return nil
	}
	
	return entry.items
}

// cacheResult stores items in cache
func (s *AISource) cacheResult(key string, items []completion.CompletionItem) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	
	s.cache[key] = cacheEntry{
		items:     items,
		timestamp: time.Now(),
	}
}

// buildPrompt creates the AI prompt
func (s *AISource) buildPrompt(ctx completion.Context) string {
	prompt := fmt.Sprintf(`Given this SQL query context, suggest completions.

Current query:
%s

Cursor is at position %d, current word being typed: "%s"
Context: %s
Available tables: %s

Provide 3-5 SQL completion suggestions. Each line should be a single completion.
Only output the completions, one per line.`, 
		ctx.Query, ctx.Cursor, ctx.Word, ctx.LinePrefix, 
		strings.Join(ctx.Tables, ", "))
	
	return prompt
}

// parseResponse parses AI response into completion items
func (s *AISource) parseResponse(response string, ctx completion.Context) []completion.CompletionItem {
	lines := strings.Split(strings.TrimSpace(response), "\n")
	items := make([]completion.CompletionItem, 0, len(lines))
	
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Remove numbering if present (1. , - , * )
		line = strings.TrimPrefix(line, fmt.Sprintf("%d. ", i+1))
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")
		line = strings.TrimSpace(line)
		
		if line == "" {
			continue
		}
		
		item := completion.CompletionItem{
			Label:      truncate(line, 40),
			InsertText: line,
			Kind:       completion.KindAI,
			Detail:     "AI Suggestion",
			Source:     s.Name(),
			Score:      20 + float64(5-i), // Decrease score for later items
			FilterText: line,
		}
		items = append(items, item)
	}
	
	return items
}

// truncate shortens text with ellipsis
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// ClearCache clears the AI cache
func (s *AISource) ClearCache() {
	s.cacheMu.Lock()
	s.cache = make(map[string]cacheEntry)
	s.cacheMu.Unlock()
}
