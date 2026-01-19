package completion

// ItemKind represents the type of completion item
type ItemKind int

const (
	KindKeyword ItemKind = iota
	KindTable
	KindColumn
	KindFunction
	KindSnippet
	KindAI
	KindHistory
)

// CompletionItem represents a single completion suggestion
type CompletionItem struct {
	Label       string   // Display text in popup
	InsertText  string   // Text to insert when selected
	Kind        ItemKind // Type of item (table, column, etc.)
	Detail      string   // Additional info (type, description)
	Source      string   // Source name for indicator icon
	Score       float64  // Relevance score for sorting
	FilterText  string   // Text to use for fuzzy filtering
}

// Context holds information about the current completion context
type Context struct {
	Query      string   // Full query text
	Cursor     int      // Cursor position in query
	Word       string   // Current word being typed
	WordStart  int      // Start position of current word
	Database   string   // Current database name
	Tables     []string // Available tables in database
	LinePrefix string   // Text before cursor on current line
}

// Source is the interface that completion sources must implement
type Source interface {
	// Name returns the source identifier
	Name() string
	
	// Complete returns completion items for the given context
	Complete(ctx Context) ([]CompletionItem, error)
	
	// Priority returns the source priority (higher = more important)
	Priority() int
}

// KindIcon returns an icon for the item kind
func (k ItemKind) Icon() string {
	switch k {
	case KindKeyword:
		return "🔤"
	case KindTable:
		return "📊"
	case KindColumn:
		return "📋"
	case KindFunction:
		return "ƒ "
	case KindSnippet:
		return "✂️"
	case KindAI:
		return "🤖"
	case KindHistory:
		return "📜"
	default:
		return "  "
	}
}

// KindName returns a readable name for the item kind
func (k ItemKind) Name() string {
	switch k {
	case KindKeyword:
		return "Keyword"
	case KindTable:
		return "Table"
	case KindColumn:
		return "Column"
	case KindFunction:
		return "Function"
	case KindSnippet:
		return "Snippet"
	case KindAI:
		return "AI"
	case KindHistory:
		return "History"
	default:
		return "Unknown"
	}
}
