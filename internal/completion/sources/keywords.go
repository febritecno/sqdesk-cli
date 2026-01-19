package sources

import (
	"strings"

	"github.com/febritecno/sqdesk/internal/completion"
)

// SQL keywords categorized by context
var sqlKeywords = map[string][]string{
	"start": {
		"SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "ALTER", "DROP",
		"TRUNCATE", "GRANT", "REVOKE", "BEGIN", "COMMIT", "ROLLBACK",
		"EXPLAIN", "ANALYZE", "SHOW", "DESCRIBE", "USE", "WITH",
	},
	"select": {
		"DISTINCT", "ALL", "TOP", "AS", "FROM", "WHERE", "AND", "OR", "NOT",
		"IN", "BETWEEN", "LIKE", "IS", "NULL", "TRUE", "FALSE",
		"ORDER", "BY", "ASC", "DESC", "LIMIT", "OFFSET",
		"GROUP", "HAVING", "UNION", "INTERSECT", "EXCEPT",
		"JOIN", "INNER", "LEFT", "RIGHT", "FULL", "OUTER", "CROSS", "ON",
		"CASE", "WHEN", "THEN", "ELSE", "END",
	},
	"functions": {
		"COUNT", "SUM", "AVG", "MIN", "MAX", "COALESCE", "NULLIF",
		"CONCAT", "SUBSTRING", "LENGTH", "UPPER", "LOWER", "TRIM",
		"NOW", "CURRENT_DATE", "CURRENT_TIME", "CURRENT_TIMESTAMP",
		"CAST", "CONVERT", "IFNULL", "NVL",
	},
	"types": {
		"INT", "INTEGER", "BIGINT", "SMALLINT", "TINYINT",
		"VARCHAR", "CHAR", "TEXT", "NVARCHAR",
		"DECIMAL", "NUMERIC", "FLOAT", "DOUBLE", "REAL",
		"DATE", "TIME", "DATETIME", "TIMESTAMP",
		"BOOLEAN", "BOOL", "BLOB", "JSON",
	},
}

// KeywordSource provides SQL keyword completions
type KeywordSource struct{}

// NewKeywordSource creates a new keyword source
func NewKeywordSource() *KeywordSource {
	return &KeywordSource{}
}

// Name returns the source name
func (s *KeywordSource) Name() string {
	return "keywords"
}

// Priority returns the source priority
func (s *KeywordSource) Priority() int {
	return 50 // Medium priority
}

// Complete returns keyword completions based on context
func (s *KeywordSource) Complete(ctx completion.Context) ([]completion.CompletionItem, error) {
	items := make([]completion.CompletionItem, 0)
	
	// Determine which keywords to suggest based on context
	categories := s.getRelevantCategories(ctx.LinePrefix)
	
	for _, category := range categories {
		keywords, ok := sqlKeywords[category]
		if !ok {
			continue
		}
		
		for _, kw := range keywords {
			item := completion.CompletionItem{
				Label:      kw,
				InsertText: kw + " ",
				Kind:       completion.KindKeyword,
				Detail:     "SQL Keyword",
				Source:     s.Name(),
				Score:      40,
				FilterText: kw,
			}
			
			// Boost if matches current word
			if ctx.Word != "" && strings.HasPrefix(strings.ToLower(kw), strings.ToLower(ctx.Word)) {
				item.Score += 30
			}
			
			items = append(items, item)
		}
	}
	
	// Add functions
	for _, fn := range sqlKeywords["functions"] {
		item := completion.CompletionItem{
			Label:      fn + "()",
			InsertText: fn + "()",
			Kind:       completion.KindFunction,
			Detail:     "SQL Function",
			Source:     s.Name(),
			Score:      35,
			FilterText: fn,
		}
		items = append(items, item)
	}
	
	return items, nil
}

// getRelevantCategories determines which keyword categories to suggest
func (s *KeywordSource) getRelevantCategories(linePrefix string) []string {
	linePrefix = strings.ToUpper(strings.TrimSpace(linePrefix))
	
	// At start of query
	if linePrefix == "" {
		return []string{"start"}
	}
	
	// Context-aware suggestions
	if strings.HasPrefix(linePrefix, "SELECT") ||
		strings.HasPrefix(linePrefix, "UPDATE") ||
		strings.HasPrefix(linePrefix, "DELETE") {
		return []string{"select"}
	}
	
	if strings.HasPrefix(linePrefix, "CREATE") ||
		strings.HasPrefix(linePrefix, "ALTER") {
		return []string{"types"}
	}
	
	return []string{"select"} // Default
}
