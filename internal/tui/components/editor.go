package components

import (
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// KeyMap defines keybindings for the editor
type KeyMap struct {
	Undo  []string
	Redo  []string
	Copy  []string
	Paste []string
	Cut   []string
	Word  []string
}

// DefaultKeyMap returns keybindings based on OS
func DefaultKeyMap() KeyMap {
	km := KeyMap{
		Undo:  []string{"ctrl+z"},
		Redo:  []string{"ctrl+y"},
		Copy:  []string{"alt+c"},
		Paste: []string{"alt+v"},
		Cut:   []string{"alt+x"},
		Word:  []string{"ctrl+w", "alt+backspace"},
	}

	if runtime.GOOS != "darwin" {
		// On non-Mac, try to support standard Ctrl keys
		// Note: Ctrl+C might still be intercepted by global quit handler
		// unless we handle it specifically in Update
		km.Copy = append(km.Copy, "ctrl+c")
		km.Paste = append(km.Paste, "ctrl+v")
		km.Cut = append(km.Cut, "ctrl+x")
	}
	
	return km
}

// EditorMode represents the current mode of the editor
type EditorMode int

const (
	ModeNormal EditorMode = iota
	ModeInsert
	ModeVisual
)

// Position represents a cursor position
type Position struct {
	Line, Col int
}

// EditorState represents a snapshot of the editor
type EditorState struct {
	Text   string
	Cursor int
}

// Editor component for SQL editing (Sublime-like)
type Editor struct {
	textarea    textarea.Model
	width       int
	height      int
	focused     bool
	styles      EditorStyles
	
	// Suggestions
	suggestion  string
	schema      map[string][]string
	
	// Selection
	selectionStart int
	selectionEnd   int
	hasSelection   bool
	
	// History
	history      []EditorState
	historyIndex int
	lastSnapshot time.Time
	
	// KeyMap
	keyMap KeyMap
	
	// Mode
	mode EditorMode
}

// EditorStyles holds styling for the editor
type EditorStyles struct {
	Normal     lipgloss.Style
	Focused    lipgloss.Style
	Title      lipgloss.Style
	LineNum    lipgloss.Style
	Keyword    lipgloss.Style
	String     lipgloss.Style
	GhostText  lipgloss.Style
	Selection  lipgloss.Style
	Suggestion lipgloss.Style
	Type       lipgloss.Style
	Function   lipgloss.Style
	Operator   lipgloss.Style
	Mode       lipgloss.Style
}

// SQL keywords for highlighting and suggestion
var sqlKeywords = []string{
	"SELECT", "FROM", "WHERE", "LIMIT", "OFFSET", "GROUP", "HAVING", "JOIN",
	"LEFT", "RIGHT", "INNER", "OUTER", "ON", "INSERT", "INTO", "VALUES",
	"UPDATE", "SET", "DELETE", "CREATE", "TABLE", "DROP", "ALTER", "INDEX",
	"VIEW", "AS", "DISTINCT", "UNION", "ALL", "CASE", "WHEN", "THEN", "ELSE", "END",
}

var sqlOperators = []string{
	"AND", "OR", "NOT", "IN", "LIKE", "IS", "BETWEEN", "EXISTS", "TRUE", "FALSE", "NULL",
	"ORDER", "BY", "ASC", "DESC",
}

var sqlTypes = []string{
	"INT", "INTEGER", "VARCHAR", "TEXT", "BOOLEAN", "DATE", "TIMESTAMP", "FLOAT", "DOUBLE", "CHAR", "BLOB",
}

var sqlFunctions = []string{
	"COUNT", "SUM", "AVG", "MIN", "MAX", "NOW", "COALESCE", "CONCAT", "SUBSTRING", "LENGTH",
}

// ... NewEditor ...

// HighlightSQL applies basic syntax highlighting to SQL
func (e Editor) HighlightSQL(sql string) string {
	result := sql
	
	// Helper to highlight words
	highlight := func(words []string, style lipgloss.Style) {
		for _, w := range words {
			re := regexp.MustCompile(`(?i)\b` + w + `\b`)
			result = re.ReplaceAllStringFunc(result, func(match string) string {
				return style.Render(strings.ToUpper(match))
			})
		}
	}
	
	highlight(sqlKeywords, e.styles.Keyword)
	highlight(sqlOperators, e.styles.Operator)
	highlight(sqlTypes, e.styles.Type)
	highlight(sqlFunctions, e.styles.Function)
	
	// Highlight strings
	strRe := regexp.MustCompile(`'[^']*'`)
	result = strRe.ReplaceAllStringFunc(result, func(match string) string {
		return e.styles.String.Render(match)
	})
	
	return result
}

// NewEditor creates a new editor component
func NewEditor(styles EditorStyles) Editor {
	ta := textarea.New()
	ta.Placeholder = "-- Write your SQL query here..."
	ta.ShowLineNumbers = true
	ta.CharLimit = 0
	ta.SetWidth(60)
	ta.SetHeight(10)
	ta.Focus()

	e := Editor{
		textarea:     ta,
		focused:      true,
		styles:       styles,
		schema:       make(map[string][]string),
		suggestion:   "",
		history:      make([]EditorState, 0),
		historyIndex: -1,
		keyMap:       DefaultKeyMap(),
		mode:         ModeNormal,
	}
	e.snapshot()
	return e
}

// SetSchema sets the database schema for suggestions
func (e *Editor) SetSchema(schema map[string][]string) {
	e.schema = schema
}

// SetSize sets the editor dimensions
func (e *Editor) SetSize(width, height int) {
	e.width = width
	e.height = height
	e.textarea.SetWidth(width - 4)
	e.textarea.SetHeight(height - 2)
}

// SetFocused sets the focus state
func (e *Editor) SetFocused(focused bool) {
	e.focused = focused
	if focused {
		e.textarea.Focus()
	} else {
		e.textarea.Blur()
	}
}

// IsFocused returns if editor is focused
func (e Editor) IsFocused() bool {
	return e.focused
}

// GetValue returns the current SQL text
func (e Editor) GetValue() string {
	return e.textarea.Value()
}

// SetValue sets the SQL text
func (e *Editor) SetValue(value string) {
	e.textarea.SetValue(value)
}

// Update handles input for the editor
func (e Editor) Update(msg tea.Msg) (Editor, tea.Cmd) {
	if !e.focused {
		return e, nil
	}

	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch e.mode {
		case ModeNormal:
			return e.updateNormal(msg)
		case ModeVisual:
			return e.updateVisual(msg)
		case ModeInsert:
			return e.updateInsert(msg)
		}
	}

	e.textarea, cmd = e.textarea.Update(msg)
	return e, cmd
}

func (e Editor) updateNormal(msg tea.KeyMsg) (Editor, tea.Cmd) {
	var cmd tea.Cmd
	key := msg.String()

	switch key {
	case "i":
		e.mode = ModeInsert
		e.hasSelection = false
		return e, nil
	case "v":
		e.mode = ModeVisual
		e.selectionStart = e.getCursorIndex()
		e.selectionEnd = e.selectionStart
		e.hasSelection = true
		return e, nil
	case "h", "left":
		e.textarea, cmd = e.textarea.Update(tea.KeyMsg{Type: tea.KeyLeft})
	case "l", "right":
		e.textarea, cmd = e.textarea.Update(tea.KeyMsg{Type: tea.KeyRight})
	case "k", "up":
		e.textarea, cmd = e.textarea.Update(tea.KeyMsg{Type: tea.KeyUp})
	case "j", "down":
		e.textarea, cmd = e.textarea.Update(tea.KeyMsg{Type: tea.KeyDown})
	case "x":
		e.snapshot()
		e.textarea, cmd = e.textarea.Update(tea.KeyMsg{Type: tea.KeyDelete})
		e.snapshot()
	case "u":
		e.undo()
	case "ctrl+r":
		e.redo()
	case "p":
		e.pasteFromClipboard()
	case "0":
		e.textarea.CursorStart()
	case "$":
		e.textarea.CursorEnd()
	}

	// Also support global shortcuts in normal mode
	if matches(key, e.keyMap.Undo) && key != "u" {
		e.undo()
	}
	if matches(key, e.keyMap.Redo) && key != "ctrl+r" {
		e.redo()
	}
	if matches(key, e.keyMap.Paste) && key != "p" {
		e.pasteFromClipboard()
	}

	return e, cmd
}

func (e Editor) updateVisual(msg tea.KeyMsg) (Editor, tea.Cmd) {
	var cmd tea.Cmd
	key := msg.String()

	switch key {
	case "esc", "ctrl+[":
		e.mode = ModeNormal
		e.hasSelection = false
		return e, nil
	case "h", "left", "l", "right", "k", "up", "j", "down":
		// Map h,j,k,l to arrow keys for textarea
		var t tea.KeyType
		switch key {
		case "h", "left": t = tea.KeyLeft
		case "l", "right": t = tea.KeyRight
		case "k", "up": t = tea.KeyUp
		case "j", "down": t = tea.KeyDown
		}
		e.textarea, cmd = e.textarea.Update(tea.KeyMsg{Type: t})
		e.selectionEnd = e.getCursorIndex()
	case "y":
		e.copyToClipboard()
		e.mode = ModeNormal
		e.hasSelection = false
	case "d", "x":
		e.cutToClipboard()
		e.mode = ModeNormal
		e.hasSelection = false
	}

	return e, cmd
}

func (e Editor) updateInsert(msg tea.KeyMsg) (Editor, tea.Cmd) {
	var cmd tea.Cmd
	key := msg.String()

	if key == "esc" || key == "ctrl+[" {
		e.mode = ModeNormal
		return e, nil
	}

	// Handle shortcuts
	if matches(key, e.keyMap.Undo) {
		e.undo()
		return e, nil
	}
	if matches(key, e.keyMap.Redo) {
		e.redo()
		return e, nil
	}
	if matches(key, e.keyMap.Copy) {
		e.copyToClipboard()
		return e, nil
	}
	if matches(key, e.keyMap.Paste) {
		e.pasteFromClipboard()
		return e, nil
	}
	if matches(key, e.keyMap.Cut) {
		e.cutToClipboard()
		return e, nil
	}
	if matches(key, e.keyMap.Word) {
		e.snapshot()
		e.textarea, cmd = e.textarea.Update(tea.KeyMsg{Type: tea.KeyBackspace, Alt: true})
		e.snapshot()
		return e, cmd
	}

	// Handle Tab to accept suggestion
	if key == "tab" && e.suggestion != "" {
		text := e.textarea.Value()
		words := strings.Fields(text)
		if len(words) > 0 {
			lastWord := words[len(words)-1]
			if strings.HasPrefix(strings.ToUpper(e.suggestion), strings.ToUpper(lastWord)) {
				completion := e.suggestion[len(lastWord):]
				if strings.ToLower(lastWord) == lastWord {
					completion = strings.ToLower(completion)
				}
				e.textarea.SetValue(text + completion)
				e.textarea.SetCursor(len(text) + len(completion))
				e.suggestion = ""
				e.snapshot()
				return e, nil
			}
		}
	}

	// Snapshot triggers
	if msg.Type == tea.KeySpace || msg.Type == tea.KeyEnter {
		e.snapshot()
	}

	e.textarea, cmd = e.textarea.Update(msg)
	e.updateSuggestion()
	return e, cmd
}

// getCursorIndex returns the linear index of the cursor
func (e Editor) getCursorIndex() int {
	lines := strings.Split(e.textarea.Value(), "\n")
	curLine := e.textarea.Line()
	charOffset := e.textarea.LineInfo().CharOffset
	
	idx := 0
	for i := 0; i < curLine && i < len(lines); i++ {
		idx += len(lines[i]) + 1 // +1 for newline
	}
	idx += charOffset
	return idx
}

// updateSuggestion updates the suggestion text
func (e *Editor) updateSuggestion() {
	text := e.textarea.Value()
	if text == "" {
		e.suggestion = ""
		return
	}
	
	words := strings.Fields(text)
	if len(words) == 0 {
		e.suggestion = ""
		return
	}
	lastWord := words[len(words)-1] // Keep original case
	upperLastWord := strings.ToUpper(lastWord)
	
	checkList := func(list []string) bool {
		for _, item := range list {
			if strings.HasPrefix(item, upperLastWord) && item != upperLastWord {
				e.suggestion = item
				return true
			}
		}
		return false
	}
	
	if checkList(sqlKeywords) { return }
	if checkList(sqlOperators) { return }
	if checkList(sqlTypes) { return }
	if checkList(sqlFunctions) { return }
	
	// Check tables
	for table := range e.schema {
		upperTable := strings.ToUpper(table)
		if strings.HasPrefix(upperTable, upperLastWord) && upperTable != upperLastWord {
			e.suggestion = table // Keep original table case
			return
		}
	}
	
	e.suggestion = ""
}

// View renders the editor
func (e Editor) View() string {
	style := e.styles.Normal
	if e.focused {
		style = e.styles.Focused
	}

	title := e.styles.Title.Render("SQL EDITOR")
	
	// Mode indicator
	modeStr := " NORMAL "
	modeStyle := lipgloss.NewStyle().Background(lipgloss.Color("4")).Foreground(lipgloss.Color("15")) // Blue for Normal
	switch e.mode {
	case ModeInsert:
		modeStr = " INSERT "
		modeStyle = lipgloss.NewStyle().Background(lipgloss.Color("2")).Foreground(lipgloss.Color("15")) // Green for Insert
	case ModeVisual:
		modeStr = " VISUAL "
		modeStyle = lipgloss.NewStyle().Background(lipgloss.Color("5")).Foreground(lipgloss.Color("15")) // Purple for Visual
	}
	modeIndicator := modeStyle.Bold(true).Render(modeStr)

	// Suggestion bar
	suggestionText := ""
	if e.suggestion != "" {
		suggestionText = "Suggest: " + e.suggestion + " (Tab)"
	}
	suggestionBar := e.styles.Suggestion.Render(suggestionText)
	
	header := lipgloss.JoinHorizontal(lipgloss.Center, title, "  ", modeIndicator, "  ", suggestionBar)
	
	content := e.textarea.View()
	
	return style.
		Width(e.width).
		Height(e.height).
		Render(header + "\n" + content)
}



// GetSelectedText returns the text to execute
func (e Editor) GetSelectedText() string {
	if e.hasSelection {
		start, end := e.selectionStart, e.selectionEnd
		// Ensure start is before end
		if start > end {
			start, end = end, start
		}
		
		val := e.textarea.Value()
		// Bounds check
		if start < 0 { start = 0 }
		if end > len(val) { end = len(val) }
		if start > len(val) { start = len(val) }
		
		if start < end {
			return val[start:end]
		}
	}
	return e.textarea.Value()
}

// snapshot saves current state to history
func (e *Editor) snapshot() {
	current := EditorState{Text: e.textarea.Value(), Cursor: e.getCursorIndex()}
	
	// Avoid duplicate snapshots
	if e.historyIndex >= 0 && e.historyIndex < len(e.history) {
		last := e.history[e.historyIndex]
		if last.Text == current.Text {
			return
		}
	}
	
	// Truncate redo history
	if e.historyIndex < len(e.history)-1 {
		e.history = e.history[:e.historyIndex+1]
	}
	
	e.history = append(e.history, current)
	e.historyIndex = len(e.history) - 1
	e.lastSnapshot = time.Now()
}

// undo reverts to previous state
func (e *Editor) undo() {
	if e.historyIndex > 0 {
		e.historyIndex--
		state := e.history[e.historyIndex]
		e.textarea.SetValue(state.Text)
		e.setCursorIndex(state.Cursor)
	}
}

// redo reverts to next state
func (e *Editor) redo() {
	if e.historyIndex < len(e.history)-1 {
		e.historyIndex++
		state := e.history[e.historyIndex]
		e.textarea.SetValue(state.Text)
		e.setCursorIndex(state.Cursor)
	}
}

// setCursorIndex sets cursor position from linear index
func (e *Editor) setCursorIndex(idx int) {
	val := e.textarea.Value()
	if idx > len(val) { idx = len(val) }
	
	// Count newlines before idx
	lines := strings.Split(val[:idx], "\n")
	line := len(lines) - 1
	col := len(lines[line])
	
	e.textarea.CursorStart()
	for i := 0; i < line; i++ {
		e.textarea.CursorDown()
	}
	e.textarea.SetCursor(col)
}

// deleteSelection deletes selected text
func (e *Editor) deleteSelection() {
	if !e.hasSelection {
		return
	}
	
	start, end := e.selectionStart, e.selectionEnd
	if start > end {
		start, end = end, start
	}
	
	val := e.textarea.Value()
	if start < 0 { start = 0 }
	if end > len(val) { end = len(val) }
	
	newVal := val[:start] + val[end:]
	e.textarea.SetValue(newVal)
	e.setCursorIndex(start)
	e.hasSelection = false
}

// copyToClipboard copies selected text
func (e *Editor) copyToClipboard() {
	text := e.GetSelectedText()
	if text != "" && e.hasSelection {
		clipboard.WriteAll(text)
	}
}

// pasteFromClipboard pastes text from clipboard
func (e *Editor) pasteFromClipboard() {
	text, err := clipboard.ReadAll()
	if err == nil && text != "" {
		if e.hasSelection {
			e.deleteSelection()
		}
		e.snapshot()
		e.textarea.InsertString(text)
		e.snapshot()
	}
}

// cutToClipboard cuts selected text
func (e *Editor) cutToClipboard() {
	if e.hasSelection {
		e.copyToClipboard()
		e.snapshot()
		e.deleteSelection()
		e.snapshot()
	}
}

// matches checks if key matches any binding
func matches(key string, bindings []string) bool {
	for _, b := range bindings {
		if key == b {
			return true
		}
	}
	return false
}
