package components

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/febritecno/sqdesk-cli/internal/config"
)

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
	keyMap config.KeyMap
	
	// Mode
	mode EditorMode
	
	// Viewport
	offsetY int
	offsetX int
	
	// Toggle Features
	showLineNumbers bool
	softWrap        bool
	rulerColumn     int
	
	// Search
	searchMode    bool
	searchQuery   string
	searchMatches []int // Indices of matches
	searchIndex   int   // Current match index
	replaceMode   bool
	replaceQuery  string
	
	// Go to Line
	gotoLineMode  bool
	gotoLineInput string

	// Mouse
	mouseDown  bool
	mouseStart int
	posX       int
	posY       int
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
	ta.ShowLineNumbers = false // We handle line numbers in custom render
	ta.CharLimit = 0
	ta.SetWidth(60)
	ta.SetHeight(10)
	ta.Focus()

	e := Editor{
		textarea:        ta,
		focused:         true,
		styles:          styles,
		schema:          make(map[string][]string),
		suggestion:      "",
		history:         make([]EditorState, 0),
		historyIndex:    -1,
		keyMap:          config.DefaultKeyMap(),
		mode:            ModeNormal,
		showLineNumbers: true,
		softWrap:        false,
		rulerColumn:     80,
	}
	e.snapshot()
	return e
}

// SetKeyMap sets the editor keymap
func (e *Editor) SetKeyMap(km config.KeyMap) {
	e.keyMap = km
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
	e.textarea.SetHeight(height - 2) // Reserve space for header
}

// SetPosition sets the absolute position of the editor
func (e *Editor) SetPosition(x, y int) {
	e.posX = x
	e.posY = y
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

// GetCursorPosition returns the cursor position as character offset
func (e Editor) GetCursorPosition() int {
	// Get current position from textarea
	value := e.textarea.Value()
	line := e.textarea.Line()
	col := e.textarea.LineInfo().ColumnOffset
	
	// Calculate character offset
	lines := strings.Split(value, "\n")
	offset := 0
	for i := 0; i < line && i < len(lines); i++ {
		offset += len(lines[i]) + 1 // +1 for newline
	}
	offset += col
	
	return offset
}

// InsertText inserts text at the current cursor position
func (e *Editor) InsertText(text string) {
	// Get current value and cursor
	value := e.textarea.Value()
	pos := e.GetCursorPosition()
	
	// Insert text at position
	if pos > len(value) {
		pos = len(value)
	}
	newValue := value[:pos] + text + value[pos:]
	e.textarea.SetValue(newValue)
	
	// Move cursor after inserted text
	e.textarea.SetCursor(pos + len(text))
}

// ReplaceCurrentWord replaces the word being typed with the given text
func (e *Editor) ReplaceCurrentWord(text string) {
	value := e.textarea.Value()
	pos := e.GetCursorPosition()
	
	if pos > len(value) {
		pos = len(value)
	}
	
	// Find word start (go backwards from cursor)
	wordStart := pos
	for wordStart > 0 {
		c := value[wordStart-1]
		if !isWordChar(c) {
			break
		}
		wordStart--
	}
	
	// Replace word with suggestion
	newValue := value[:wordStart] + text + value[pos:]
	e.textarea.SetValue(newValue)
	
	// Move cursor after inserted text
	e.textarea.SetCursor(wordStart + len(text))
}

// isWordChar checks if character is part of a word
func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// selectAll selects the entire document
func (e *Editor) selectAll() {
	e.mode = ModeVisual
	e.selectionStart = 0
	e.selectionEnd = len(e.textarea.Value())
	e.hasSelection = true
	e.textarea.SetCursor(len(e.textarea.Value()))
}

// hasActiveSelection returns true if there's a non-empty selection
func (e Editor) hasActiveSelection() bool {
	if !e.hasSelection {
		return false
	}
	start, end := e.selectionStart, e.selectionEnd
	if start > end {
		start, end = end, start
	}
	return start != end
}

// clearSelection clears the current selection
func (e *Editor) clearSelection() {
	e.hasSelection = false
	e.selectionStart = 0
	e.selectionEnd = 0
}

// Update handles input for the editor
func (e Editor) Update(msg tea.Msg) (Editor, tea.Cmd) {
	if !e.focused {
		return e, nil
	}

	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.MouseMsg:
		return e.handleMouse(msg)
	case tea.KeyMsg:
		// Go to Line mode handling
		if e.gotoLineMode {
			return e.updateGotoLineInput(msg)
		}
		
		// Search mode handling
		if e.searchMode {
			return e.updateSearchInput(msg)
		}
		
		// Global shortcuts (work in all modes)
		key := msg.String()
		if key == "ctrl+f" {
			e.startSearch()
			return e, nil
		}
		if key == "ctrl+h" {
			e.startReplace()
			return e, nil
		}
		if key == "f3" {
			e.findNext()
			return e, nil
		}
		if key == "shift+f3" {
			e.findPrev()
			return e, nil
		}
		if key == "ctrl+g" {
			e.startGotoLine()
			return e, nil
		}
		
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

	// Handle Select All
	if matches(key, e.keyMap.SelectAll) {
		e.selectAll()
		return e, nil
	}

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
	case "alt+up":
		e.moveLineUp()
	case "alt+down":
		e.moveLineDown()
	case "ctrl+d":
		e.duplicateLine()
	case "ctrl+k":
		e.deleteLine()
	case "ctrl+/":
		e.toggleComment()
	case "alt+z":
		e.toggleSoftWrap()
	case "ctrl+l":
		e.toggleLineNumbers()
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
	case "tab":
		e.indentSelection()
	case "shift+tab":
		e.outdentSelection()
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

	// Handle Select All
	if matches(key, e.keyMap.SelectAll) {
		e.selectAll()
		return e, nil
	}

	// Handle Shift+Arrow keys for selection
	switch key {
	case "shift+left", "shift+right", "shift+up", "shift+down":
		if !e.hasSelection {
			// Start new selection
			e.selectionStart = e.getCursorIndex()
			e.hasSelection = true
		}
		
		// Move cursor
		var keyType tea.KeyType
		switch key {
		case "shift+left":
			keyType = tea.KeyLeft
		case "shift+right":
			keyType = tea.KeyRight
		case "shift+up":
			keyType = tea.KeyUp
		case "shift+down":
			keyType = tea.KeyDown
		}
		e.textarea, cmd = e.textarea.Update(tea.KeyMsg{Type: keyType})
		e.selectionEnd = e.getCursorIndex()
		return e, cmd
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
		
		// Auto-indent
		if msg.Type == tea.KeyEnter {
			lines := strings.Split(e.textarea.Value(), "\n")
			curLine := e.textarea.Line()
			if curLine > 0 && curLine < len(lines) {
				prevLine := lines[curLine-1]
				indent := ""
				for _, char := range prevLine {
					if char == ' ' || char == '\t' {
						indent += string(char)
					} else {
						break
					}
				}
				if indent != "" {
					e.textarea.InsertString(indent)
				}
			}
		}
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
	
	content := e.render()
	
	// Search bar
	searchBar := ""
	if e.searchMode {
		matchInfo := ""
		if len(e.searchMatches) > 0 {
			matchInfo = fmt.Sprintf(" (%d/%d)", e.searchIndex+1, len(e.searchMatches))
		}
		searchBar = lipgloss.NewStyle().
			Background(lipgloss.Color("8")).
			Foreground(lipgloss.Color("15")).
			Padding(0, 1).
			Render(fmt.Sprintf("Find: %s%s", e.searchQuery, matchInfo))
		
		if e.replaceMode {
			replaceBar := lipgloss.NewStyle().
				Background(lipgloss.Color("8")).
				Foreground(lipgloss.Color("15")).
				Padding(0, 1).
				Render(fmt.Sprintf("Replace: %s", e.replaceQuery))
			searchBar += "\n" + replaceBar
		}
		searchBar += "\n"
	}
	
	// Go to Line bar
	gotoLineBar := ""
	if e.gotoLineMode {
		lines := strings.Split(e.textarea.Value(), "\n")
		gotoLineBar = lipgloss.NewStyle().
			Background(lipgloss.Color("3")).
			Foreground(lipgloss.Color("0")).
			Padding(0, 1).
			Render(fmt.Sprintf("Go to line (1-%d): %s", len(lines), e.gotoLineInput))
		gotoLineBar += "\n"
	}
	
	return style.
		Width(e.width).
		Height(e.height).
		Render(header + "\n" + searchBar + gotoLineBar + content)
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

// moveLineUp moves current line up
func (e *Editor) moveLineUp() {
	lines := strings.Split(e.textarea.Value(), "\n")
	curLine := e.textarea.Line()
	if curLine > 0 {
		lines[curLine], lines[curLine-1] = lines[curLine-1], lines[curLine]
		e.textarea.SetValue(strings.Join(lines, "\n"))
		e.textarea.CursorUp()
		e.snapshot()
	}
}

// moveLineDown moves current line down
func (e *Editor) moveLineDown() {
	lines := strings.Split(e.textarea.Value(), "\n")
	curLine := e.textarea.Line()
	if curLine < len(lines)-1 {
		lines[curLine], lines[curLine+1] = lines[curLine+1], lines[curLine]
		e.textarea.SetValue(strings.Join(lines, "\n"))
		e.textarea.CursorDown()
		e.snapshot()
	}
}

// duplicateLine duplicates current line
func (e *Editor) duplicateLine() {
	lines := strings.Split(e.textarea.Value(), "\n")
	curLine := e.textarea.Line()
	if curLine < len(lines) {
		newLine := lines[curLine]
		lines = append(lines[:curLine+1], append([]string{newLine}, lines[curLine+1:]...)...)
		e.textarea.SetValue(strings.Join(lines, "\n"))
		e.textarea.CursorDown()
		e.snapshot()
	}
}

// deleteLine deletes current line
func (e *Editor) deleteLine() {
	lines := strings.Split(e.textarea.Value(), "\n")
	curLine := e.textarea.Line()
	
	if len(lines) > 0 {
		lines = append(lines[:curLine], lines[curLine+1:]...)
		if len(lines) == 0 { lines = []string{""} }
		e.textarea.SetValue(strings.Join(lines, "\n"))
		if curLine >= len(lines) && curLine > 0 {
			e.textarea.CursorUp()
		}
		e.snapshot()
	}
}

// getLineCol returns line and col for index
func (e Editor) getLineCol(idx int) (int, int) {
	lines := strings.Split(e.textarea.Value(), "\n")
	current := 0
	for i, line := range lines {
		next := current + len(line) + 1
		if idx < next {
			return i, idx - current
		}
		current = next
	}
	if len(lines) == 0 { return 0, 0 }
	return len(lines) - 1, len(lines[len(lines)-1])
}

// indentSelection adds indentation to selected lines
func (e *Editor) indentSelection() {
	if !e.hasSelection { return }
	lines := strings.Split(e.textarea.Value(), "\n")
	
	startIdx, endIdx := e.selectionStart, e.selectionEnd
	if startIdx > endIdx { startIdx, endIdx = endIdx, startIdx }
	
	startLine, _ := e.getLineCol(startIdx)
	endLine, _ := e.getLineCol(endIdx)
	
	for i := startLine; i <= endLine && i < len(lines); i++ {
		lines[i] = "  " + lines[i]
	}
	
	e.textarea.SetValue(strings.Join(lines, "\n"))
	e.selectionEnd += (endLine - startLine + 1) * 2
	e.snapshot()
}

// outdentSelection removes indentation from selected lines
func (e *Editor) outdentSelection() {
	if !e.hasSelection { return }
	lines := strings.Split(e.textarea.Value(), "\n")
	
	startIdx, endIdx := e.selectionStart, e.selectionEnd
	if startIdx > endIdx { startIdx, endIdx = endIdx, startIdx }
	
	startLine, _ := e.getLineCol(startIdx)
	endLine, _ := e.getLineCol(endIdx)
	
	removedChars := 0
	for i := startLine; i <= endLine && i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "  ") {
			lines[i] = lines[i][2:]
			removedChars += 2
		} else if strings.HasPrefix(lines[i], " ") {
			lines[i] = lines[i][1:]
			removedChars += 1
		}
	}
	
	e.textarea.SetValue(strings.Join(lines, "\n"))
	e.selectionEnd -= removedChars
	e.snapshot()
}

// toggleComment toggles SQL comment on current line or selected lines
func (e *Editor) toggleComment() {
	lines := strings.Split(e.textarea.Value(), "\n")
	
	startLine := e.textarea.Line()
	endLine := startLine
	
	if e.hasSelection {
		startIdx, endIdx := e.selectionStart, e.selectionEnd
		if startIdx > endIdx { startIdx, endIdx = endIdx, startIdx }
		startLine, _ = e.getLineCol(startIdx)
		endLine, _ = e.getLineCol(endIdx)
	}
	
	// Check if all lines are commented
	allCommented := true
	for i := startLine; i <= endLine && i < len(lines); i++ {
		trimmed := strings.TrimLeft(lines[i], " \t")
		if !strings.HasPrefix(trimmed, "-- ") && !strings.HasPrefix(trimmed, "--") {
			allCommented = false
			break
		}
	}
	
	for i := startLine; i <= endLine && i < len(lines); i++ {
		if allCommented {
			// Remove comment
			if idx := strings.Index(lines[i], "-- "); idx != -1 {
				lines[i] = lines[i][:idx] + lines[i][idx+3:]
			} else if idx := strings.Index(lines[i], "--"); idx != -1 {
				lines[i] = lines[i][:idx] + lines[i][idx+2:]
			}
		} else {
			// Add comment
			lines[i] = "-- " + lines[i]
		}
	}
	
	e.textarea.SetValue(strings.Join(lines, "\n"))
	e.snapshot()
}

// toggleSoftWrap toggles soft wrap mode
func (e *Editor) toggleSoftWrap() {
	e.softWrap = !e.softWrap
}

// toggleLineNumbers toggles line number display
func (e *Editor) toggleLineNumbers() {
	e.showLineNumbers = !e.showLineNumbers
}

// startSearch enters search mode
func (e *Editor) startSearch() {
	e.searchMode = true
	e.searchQuery = ""
	e.searchMatches = nil
	e.searchIndex = 0
}

// startReplace enters replace mode
func (e *Editor) startReplace() {
	e.searchMode = true
	e.replaceMode = true
	e.searchQuery = ""
	e.replaceQuery = ""
	e.searchMatches = nil
	e.searchIndex = 0
}

// cancelSearch exits search mode
func (e *Editor) cancelSearch() {
	e.searchMode = false
	e.replaceMode = false
}

// updateSearchInput handles key input in search mode
func (e Editor) updateSearchInput(msg tea.KeyMsg) (Editor, tea.Cmd) {
	key := msg.String()
	
	switch key {
	case "esc":
		e.cancelSearch()
		return e, nil
	case "enter":
		if e.replaceMode {
			e.doReplace()
		} else {
			e.findNext()
		}
		return e, nil
	case "f3":
		e.findNext()
		return e, nil
	case "shift+f3":
		e.findPrev()
		return e, nil
	case "ctrl+enter":
		if e.replaceMode {
			e.replaceAll()
		}
		return e, nil
	case "backspace":
		if len(e.searchQuery) > 0 {
			e.searchQuery = e.searchQuery[:len(e.searchQuery)-1]
			e.refreshSearchMatches()
		}
		return e, nil
	case "tab":
		// Switch focus between search and replace fields (if in replace mode)
		// For simplicity, we handle this by just moving between fields conceptually
		return e, nil
	default:
		// Add character to search query
		if len(key) == 1 && key >= " " && key <= "~" {
			e.searchQuery += key
			e.refreshSearchMatches()
		}
		return e, nil
	}
}

// refreshSearchMatches updates search matches
func (e *Editor) refreshSearchMatches() {
	if e.searchQuery == "" {
		e.searchMatches = nil
		return
	}
	
	text := e.textarea.Value()
	query := e.searchQuery
	e.searchMatches = nil
	
	idx := 0
	for {
		found := strings.Index(text[idx:], query)
		if found == -1 {
			break
		}
		e.searchMatches = append(e.searchMatches, idx+found)
		idx += found + 1
	}
	
	if e.searchIndex >= len(e.searchMatches) {
		e.searchIndex = 0
	}
}

// findNext moves to next match
func (e *Editor) findNext() {
	if len(e.searchMatches) == 0 {
		return
	}
	e.searchIndex = (e.searchIndex + 1) % len(e.searchMatches)
	e.setCursorIndex(e.searchMatches[e.searchIndex])
	e.updateViewport()
}

// findPrev moves to previous match
func (e *Editor) findPrev() {
	if len(e.searchMatches) == 0 {
		return
	}
	e.searchIndex--
	if e.searchIndex < 0 {
		e.searchIndex = len(e.searchMatches) - 1
	}
	e.setCursorIndex(e.searchMatches[e.searchIndex])
	e.updateViewport()
}

// doReplace replaces current match
func (e *Editor) doReplace() {
	if len(e.searchMatches) == 0 || e.replaceQuery == "" {
		return
	}
	
	matchIdx := e.searchMatches[e.searchIndex]
	text := e.textarea.Value()
	newText := text[:matchIdx] + e.replaceQuery + text[matchIdx+len(e.searchQuery):]
	e.textarea.SetValue(newText)
	e.snapshot()
	e.refreshSearchMatches()
}

// replaceAll replaces all matches
func (e *Editor) replaceAll() {
	if e.searchQuery == "" || e.replaceQuery == "" {
		return
	}
	
	text := e.textarea.Value()
	newText := strings.ReplaceAll(text, e.searchQuery, e.replaceQuery)
	e.textarea.SetValue(newText)
	e.snapshot()
	e.refreshSearchMatches()
}

// startGotoLine enters go to line mode
func (e *Editor) startGotoLine() {
	e.gotoLineMode = true
	e.gotoLineInput = ""
}

// cancelGotoLine exits go to line mode
func (e *Editor) cancelGotoLine() {
	e.gotoLineMode = false
	e.gotoLineInput = ""
}

// doGotoLine jumps to the specified line
func (e *Editor) doGotoLine() {
	lineNum, err := strconv.Atoi(e.gotoLineInput)
	if err != nil || lineNum < 1 {
		e.cancelGotoLine()
		return
	}
	
	lines := strings.Split(e.textarea.Value(), "\n")
	if lineNum > len(lines) {
		lineNum = len(lines)
	}
	
	// Calculate index of line start
	idx := 0
	for i := 0; i < lineNum-1 && i < len(lines); i++ {
		idx += len(lines[i]) + 1
	}
	
	e.setCursorIndex(idx)
	e.updateViewport()
	e.cancelGotoLine()
}

// updateGotoLineInput handles key input in go to line mode
func (e Editor) updateGotoLineInput(msg tea.KeyMsg) (Editor, tea.Cmd) {
	key := msg.String()
	
	switch key {
	case "esc":
		e.cancelGotoLine()
		return e, nil
	case "enter":
		e.doGotoLine()
		return e, nil
	case "backspace":
		if len(e.gotoLineInput) > 0 {
			e.gotoLineInput = e.gotoLineInput[:len(e.gotoLineInput)-1]
		}
		return e, nil
	default:
		// Only accept digits
		if len(key) == 1 && key >= "0" && key <= "9" {
			e.gotoLineInput += key
		}
		return e, nil
	}
}

// handleMouse handles mouse events
func (e Editor) handleMouse(msg tea.MouseMsg) (Editor, tea.Cmd) {
	// Calculate dynamic header height
	headerHeight := 1 // Title bar
	if e.searchMode {
		headerHeight++
		if e.replaceMode {
			headerHeight++
		}
	}
	if e.gotoLineMode {
		headerHeight++
	}
	
	// Check if click is within content area
	contentY := e.posY + headerHeight
	if msg.Y < contentY || msg.Y >= contentY + (e.height - headerHeight) {
		return e, nil
	}
	
	// Calculate index
	idx := e.getIndexFromCoords(msg.X, msg.Y, headerHeight)
	if idx == -1 {
		return e, nil
	}
	
	switch msg.Type {
	case tea.MouseLeft:
		if !e.mouseDown {
			e.mouseDown = true
			e.mouseStart = idx
			e.setCursorIndex(idx)
			e.hasSelection = false
			
			// If clicking, ensure we are in a mode that supports selection or switch to Visual if dragging starts
			// For now, just move cursor. If drag happens, we switch to Visual.
		} else {
			// Dragging
			if idx != e.mouseStart {
				e.mode = ModeVisual
				e.selectionStart = e.mouseStart
				e.selectionEnd = idx
				e.hasSelection = true
				e.setCursorIndex(idx)
			}
		}
	case tea.MouseRelease:
		e.mouseDown = false
		if e.hasSelection {
			// Keep selection
		}
	case tea.MouseWheelUp:
		e.textarea.CursorUp()
		e.updateViewport()
	case tea.MouseWheelDown:
		e.textarea.CursorDown()
		e.updateViewport()
	}
	
	return e, nil
}

// getIndexFromCoords converts mouse coordinates to text index
func (e Editor) getIndexFromCoords(x, y, headerHeight int) int {
	contentY := e.posY + headerHeight
	relY := y - contentY
	lineIdx := e.offsetY + relY
	
	lines := strings.Split(e.textarea.Value(), "\n")
	if lineIdx < 0 || lineIdx >= len(lines) {
		return -1
	}
	
	// Line number width is 5 ("%4d ")
	contentX := e.posX + 5
	relX := x - contentX
	if relX < 0 { relX = 0 }
	
	line := lines[lineIdx]
	if relX > len(line) {
		relX = len(line)
	}
	
	// Calculate global index
	idx := 0
	for i := 0; i < lineIdx; i++ {
		idx += len(lines[i]) + 1 // +1 for newline
	}
	idx += relX
	
	return idx
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

// updateViewport adjusts the viewport to keep the cursor visible
func (e *Editor) updateViewport() {
	cursorLine := e.textarea.Line()
	viewportHeight := e.height - 2
	
	if cursorLine < e.offsetY {
		e.offsetY = cursorLine
	} else if cursorLine >= e.offsetY+viewportHeight {
		e.offsetY = cursorLine - viewportHeight + 1
	}
	
	if e.offsetY < 0 { e.offsetY = 0 }
}

// render renders the editor content with custom highlighting
func (e Editor) render() string {
	var view strings.Builder
	lines := strings.Split(e.textarea.Value(), "\n")
	viewportHeight := e.height - 2
	
	// Ensure offsets are valid
	if e.offsetY > len(lines) { e.offsetY = len(lines) - 1 }
	if e.offsetY < 0 { e.offsetY = 0 }
	
	startLine := e.offsetY
	endLine := startLine + viewportHeight
	if endLine > len(lines) { endLine = len(lines) }
	
	// Calculate global index for startLine
	currentIdx := 0
	for i := 0; i < startLine; i++ {
		currentIdx += len(lines[i]) + 1
	}
	
	selStart, selEnd := e.selectionStart, e.selectionEnd
	if selStart > selEnd { selStart, selEnd = selEnd, selStart }
	
	cursorLine := e.textarea.Line()
	cursorCol := e.textarea.LineInfo().ColumnOffset
	
	for i := startLine; i < endLine; i++ {
		line := lines[i]
		lineLen := len(line)
		lineEndIdx := currentIdx + lineLen
		
		// Line Number (conditional)
		if e.showLineNumbers {
			lineNumStyle := e.styles.LineNum
			if i == cursorLine {
				lineNumStyle = e.styles.LineNum.Copy().Foreground(lipgloss.Color("15")).Bold(true)
			}
			view.WriteString(lineNumStyle.Render(fmt.Sprintf("%4d ", i+1)))
		}
		
		// Determine cuts for segments
		cCol := -1
		if i == cursorLine {
			cCol = cursorCol
			if cCol > lineLen { cCol = lineLen }
		}
		
		sStart, sEnd := -1, -1
		if e.hasSelection && selEnd > currentIdx && selStart < lineEndIdx {
			sStart = selStart - currentIdx
			sEnd = selEnd - currentIdx
			if sStart < 0 { sStart = 0 }
			if sEnd > lineLen { sEnd = lineLen }
		}
		
		cuts := []int{0, lineLen}
		if sStart != -1 {
			cuts = append(cuts, sStart, sEnd)
		}
		if cCol != -1 {
			cuts = append(cuts, cCol, cCol+1)
		}
		
		// Sort and unique
		sort.Ints(cuts)
		uniqueCuts := make([]int, 0, len(cuts))
		if len(cuts) > 0 {
			uniqueCuts = append(uniqueCuts, cuts[0])
			for j := 1; j < len(cuts); j++ {
				if cuts[j] != cuts[j-1] {
					uniqueCuts = append(uniqueCuts, cuts[j])
				}
			}
		}
		cuts = uniqueCuts
		
		// Render segments
		for k := 0; k < len(cuts)-1; k++ {
			p1, p2 := cuts[k], cuts[k+1]
			if p1 >= p2 { continue }
			if p1 >= lineLen && cCol != p1 { continue }
			
			segText := ""
			if p1 < lineLen {
				end := p2
				if end > lineLen { end = lineLen }
				segText = line[p1:end]
			} else if p1 == lineLen && cCol == p1 {
				segText = " "
			}
			
			isSel := (sStart != -1 && p1 >= sStart && p1 < sEnd)
			isCur := (cCol != -1 && p1 == cCol)
			
			if isCur {
				view.WriteString(lipgloss.NewStyle().Reverse(true).Render(segText))
			} else if isSel {
				view.WriteString(e.styles.Selection.Render(segText))
			} else {
				view.WriteString(e.HighlightSQL(segText))
			}
		}
		
		// If cursor is at end of line and line is empty or we didn't process it
		if cCol == lineLen && lineLen == 0 {
             view.WriteString(lipgloss.NewStyle().Reverse(true).Render(" "))
        }
		
		view.WriteString("\n")
		currentIdx += lineLen + 1
	}
	
	// Fill empty lines
	for i := endLine; i < startLine + viewportHeight; i++ {
		view.WriteString(e.styles.LineNum.Render("~") + "\n")
	}
	
	return view.String()
}
