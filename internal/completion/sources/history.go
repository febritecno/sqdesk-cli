package sources

import (
	"strings"

	"github.com/febritecno/sqdesk-cli/internal/completion"
)

// HistorySource provides completions from query history
type HistorySource struct {
	history []string // Most recent first
	maxSize int
}

// NewHistorySource creates a new history source
func NewHistorySource() *HistorySource {
	return &HistorySource{
		history: make([]string, 0),
		maxSize: 100,
	}
}

// Name returns the source name
func (s *HistorySource) Name() string {
	return "history"
}

// Priority returns the source priority
func (s *HistorySource) Priority() int {
	return 70 // Higher than keywords, lower than schema
}

// AddQuery adds a query to history
func (s *HistorySource) AddQuery(query string) {
	query = strings.TrimSpace(query)
	if query == "" {
		return
	}
	
	// Remove if already exists
	for i, q := range s.history {
		if q == query {
			s.history = append(s.history[:i], s.history[i+1:]...)
			break
		}
	}
	
	// Add to front
	s.history = append([]string{query}, s.history...)
	
	// Trim to max size
	if len(s.history) > s.maxSize {
		s.history = s.history[:s.maxSize]
	}
}

// Complete returns history-based completions
func (s *HistorySource) Complete(ctx completion.Context) ([]completion.CompletionItem, error) {
	items := make([]completion.CompletionItem, 0)
	
	prefix := strings.ToUpper(ctx.LinePrefix)
	
	for i, query := range s.history {
		// Match queries that start similar
		queryUpper := strings.ToUpper(query)
		if prefix != "" && !strings.HasPrefix(queryUpper, prefix) {
			continue
		}
		
		// Create item
		item := completion.CompletionItem{
			Label:      truncateHistory(query, 50),
			InsertText: query,
			Kind:       completion.KindHistory,
			Detail:     "From history",
			Source:     s.Name(),
			Score:      60 - float64(i), // More recent = higher score
			FilterText: query,
		}
		items = append(items, item)
		
		// Limit results
		if len(items) >= 5 {
			break
		}
	}
	
	return items, nil
}

// truncateHistory shortens query for display
func truncateHistory(s string, max int) string {
	// Remove newlines
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	
	// Collapse multiple spaces
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// Clear clears the history
func (s *HistorySource) Clear() {
	s.history = make([]string, 0)
}

// GetHistory returns the query history
func (s *HistorySource) GetHistory() []string {
	return s.history
}
