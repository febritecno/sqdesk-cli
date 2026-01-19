package sources

import (
	"strings"

	"github.com/febritecno/sqdesk/internal/completion"
)

// SchemaSource provides table and column completions from database schema
type SchemaSource struct {
	tables  []TableInfo
	columns map[string][]ColumnInfo // table name -> columns
}

// TableInfo holds table metadata
type TableInfo struct {
	Name    string
	Schema  string
	Comment string
}

// ColumnInfo holds column metadata
type ColumnInfo struct {
	Name      string
	Type      string
	Nullable  bool
	IsPrimary bool
	Comment   string
}

// NewSchemaSource creates a new schema source
func NewSchemaSource() *SchemaSource {
	return &SchemaSource{
		tables:  make([]TableInfo, 0),
		columns: make(map[string][]ColumnInfo),
	}
}

// Name returns the source name
func (s *SchemaSource) Name() string {
	return "schema"
}

// Priority returns the source priority
func (s *SchemaSource) Priority() int {
	return 100 // High priority - schema is most relevant
}

// SetTables sets the available tables
func (s *SchemaSource) SetTables(tables []TableInfo) {
	s.tables = tables
}

// SetColumns sets the columns for a table
func (s *SchemaSource) SetColumns(tableName string, columns []ColumnInfo) {
	s.columns[tableName] = columns
}

// LoadFromStrings loads tables from simple string slice
func (s *SchemaSource) LoadFromStrings(tableNames []string) {
	s.tables = make([]TableInfo, len(tableNames))
	for i, name := range tableNames {
		s.tables[i] = TableInfo{Name: name}
	}
}

// Complete returns schema-based completions
func (s *SchemaSource) Complete(ctx completion.Context) ([]completion.CompletionItem, error) {
	items := make([]completion.CompletionItem, 0)
	
	// Determine if we should suggest tables or columns
	suggestTables := s.shouldSuggestTables(ctx.LinePrefix)
	suggestColumns := s.shouldSuggestColumns(ctx.LinePrefix)
	
	if suggestTables {
		items = append(items, s.getTableItems(ctx)...)
	}
	
	if suggestColumns {
		items = append(items, s.getColumnItems(ctx)...)
	}
	
	return items, nil
}

// shouldSuggestTables checks if tables should be suggested
func (s *SchemaSource) shouldSuggestTables(linePrefix string) bool {
	linePrefix = strings.ToUpper(linePrefix)
	
	// After FROM, JOIN, INTO, UPDATE, etc.
	triggers := []string{"FROM", "JOIN", "INTO", "UPDATE", "TABLE", "TRUNCATE"}
	for _, trigger := range triggers {
		if strings.HasSuffix(linePrefix, trigger) || strings.HasSuffix(linePrefix, trigger+" ") {
			return true
		}
	}
	
	// At start or after comma in FROM clause
	if strings.Contains(linePrefix, "FROM") {
		return true
	}
	
	return false
}

// shouldSuggestColumns checks if columns should be suggested
func (s *SchemaSource) shouldSuggestColumns(linePrefix string) bool {
	linePrefix = strings.ToUpper(linePrefix)
	
	// After SELECT, WHERE, SET, ORDER BY, GROUP BY
	triggers := []string{"SELECT", "WHERE", "AND", "OR", "SET", "ORDER BY", "GROUP BY", "HAVING"}
	for _, trigger := range triggers {
		if strings.HasSuffix(linePrefix, trigger) || strings.HasSuffix(linePrefix, trigger+" ") {
			return true
		}
	}
	
	// After dot (table.column)
	if strings.HasSuffix(linePrefix, ".") {
		return true
	}
	
	return false
}

// getTableItems returns completion items for tables
func (s *SchemaSource) getTableItems(ctx completion.Context) []completion.CompletionItem {
	items := make([]completion.CompletionItem, 0, len(s.tables))
	
	for _, table := range s.tables {
		detail := "Table"
		if table.Schema != "" {
			detail = table.Schema + "." + table.Name
		}
		if table.Comment != "" {
			detail += " - " + table.Comment
		}
		
		item := completion.CompletionItem{
			Label:      table.Name,
			InsertText: table.Name,
			Kind:       completion.KindTable,
			Detail:     detail,
			Source:     s.Name(),
			Score:      90,
			FilterText: table.Name,
		}
		items = append(items, item)
	}
	
	return items
}

// getColumnItems returns completion items for columns
func (s *SchemaSource) getColumnItems(ctx completion.Context) []completion.CompletionItem {
	items := make([]completion.CompletionItem, 0)
	
	// Suggest columns from all known tables
	for tableName, columns := range s.columns {
		for _, col := range columns {
			detail := col.Type
			if col.IsPrimary {
				detail += " PRIMARY KEY"
			}
			if !col.Nullable {
				detail += " NOT NULL"
			}
			
			item := completion.CompletionItem{
				Label:      col.Name,
				InsertText: col.Name,
				Kind:       completion.KindColumn,
				Detail:     tableName + "." + col.Name + " (" + detail + ")",
				Source:     s.Name(),
				Score:      85,
				FilterText: col.Name,
			}
			items = append(items, item)
		}
	}
	
	return items
}

// Clear clears all schema data
func (s *SchemaSource) Clear() {
	s.tables = make([]TableInfo, 0)
	s.columns = make(map[string][]ColumnInfo)
}
