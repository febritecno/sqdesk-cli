package completion

import (
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
)

// Engine manages completion sources and provides completions
type Engine struct {
	sources     []Source
	mu          sync.RWMutex
	debounceMs  int
	lastTrigger time.Time
	cache       map[string][]CompletionItem
	cacheMu     sync.RWMutex
}

// NewEngine creates a new completion engine
func NewEngine() *Engine {
	return &Engine{
		sources:    make([]Source, 0),
		debounceMs: 100,
		cache:      make(map[string][]CompletionItem),
	}
}

// RegisterSource adds a completion source to the engine
func (e *Engine) RegisterSource(source Source) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.sources = append(e.sources, source)
	
	// Sort sources by priority (highest first)
	sort.Slice(e.sources, func(i, j int) bool {
		return e.sources[i].Priority() > e.sources[j].Priority()
	})
}

// Complete returns completion items for the given query and cursor position
func (e *Engine) Complete(query string, cursor int, database string, tables []string) []CompletionItem {
	ctx := e.buildContext(query, cursor, database, tables)
	
	// Check cache
	cacheKey := ctx.Word + "|" + ctx.LinePrefix
	e.cacheMu.RLock()
	if cached, ok := e.cache[cacheKey]; ok {
		e.cacheMu.RUnlock()
		return cached
	}
	e.cacheMu.RUnlock()
	
	// Collect items from all sources
	var allItems []CompletionItem
	var wg sync.WaitGroup
	var itemsMu sync.Mutex
	
	e.mu.RLock()
	sources := e.sources
	e.mu.RUnlock()
	
	for _, source := range sources {
		wg.Add(1)
		go func(s Source) {
			defer wg.Done()
			items, err := s.Complete(ctx)
			if err != nil {
				return
			}
			itemsMu.Lock()
			allItems = append(allItems, items...)
			itemsMu.Unlock()
		}(source)
	}
	
	wg.Wait()
	
	// Filter by current word
	if ctx.Word != "" {
		allItems = e.filterItems(allItems, ctx.Word)
	}
	
	// Sort by score
	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].Score > allItems[j].Score
	})
	
	// Limit results
	if len(allItems) > 20 {
		allItems = allItems[:20]
	}
	
	// Cache results
	e.cacheMu.Lock()
	e.cache[cacheKey] = allItems
	e.cacheMu.Unlock()
	
	return allItems
}

// buildContext creates a completion context from the query
func (e *Engine) buildContext(query string, cursor int, database string, tables []string) Context {
	// Extract current word
	word, wordStart := extractWord(query, cursor)
	
	// Extract line prefix
	lineStart := strings.LastIndex(query[:cursor], "\n")
	if lineStart == -1 {
		lineStart = 0
	} else {
		lineStart++
	}
	linePrefix := strings.ToUpper(strings.TrimSpace(query[lineStart:cursor]))
	
	return Context{
		Query:      query,
		Cursor:     cursor,
		Word:       word,
		WordStart:  wordStart,
		Database:   database,
		Tables:     tables,
		LinePrefix: linePrefix,
	}
}

// extractWord extracts the word being typed at cursor position
func extractWord(query string, cursor int) (string, int) {
	if cursor > len(query) {
		cursor = len(query)
	}
	
	// Find word start
	start := cursor
	for start > 0 {
		r := rune(query[start-1])
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			break
		}
		start--
	}
	
	return query[start:cursor], start
}

// filterItems filters items by fuzzy matching the prefix
func (e *Engine) filterItems(items []CompletionItem, prefix string) []CompletionItem {
	prefix = strings.ToLower(prefix)
	result := make([]CompletionItem, 0)
	
	for _, item := range items {
		filterText := item.FilterText
		if filterText == "" {
			filterText = item.Label
		}
		filterText = strings.ToLower(filterText)
		
		// Simple prefix match or contains
		if strings.HasPrefix(filterText, prefix) {
			item.Score += 100 // Boost exact prefix matches
			result = append(result, item)
		} else if strings.Contains(filterText, prefix) {
			item.Score += 50 // Partial match
			result = append(result, item)
		}
	}
	
	return result
}

// ClearCache clears the completion cache
func (e *Engine) ClearCache() {
	e.cacheMu.Lock()
	e.cache = make(map[string][]CompletionItem)
	e.cacheMu.Unlock()
}

// SetDebounce sets the debounce time in milliseconds
func (e *Engine) SetDebounce(ms int) {
	e.debounceMs = ms
}
