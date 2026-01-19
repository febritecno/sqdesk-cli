package components

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/guptarohit/asciigraph"
)

// ViewMode determines how results are displayed
type ViewMode int

const (
	ViewTable ViewMode = iota
	ViewChartBar
	ViewChartLine
	ViewChartPie
)

// Results component for displaying query results
type Results struct {
	table     table.Model
	columns   []string
	rows      []map[string]interface{}
	width     int
	height    int
	focused   bool
	styles    ResultsStyles
	rowCount  int
	message   string
	isError   bool
	page      int
	pageSize  int
	viewMode  ViewMode
}

// ResultsStyles holds styling for the results
type ResultsStyles struct {
	Normal      lipgloss.Style
	Focused     lipgloss.Style
	Title       lipgloss.Style
	Header      lipgloss.Style
	Cell        lipgloss.Style
	SelectedRow lipgloss.Style
	Error       lipgloss.Style
	Info        lipgloss.Style
}

// NewResults creates a new results component
func NewResults(styles ResultsStyles) Results {
	columns := []table.Column{}
	rows := []table.Row{}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(false),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = styles.Header
	s.Selected = styles.SelectedRow
	s.Cell = styles.Cell
	t.SetStyles(s)

	return Results{
		table:    t,
		focused:  false,
		styles:   styles,
		page:     0,
		pageSize: 100,
	}
}

// SetSize sets the results dimensions
func (r *Results) SetSize(width, height int) {
	r.width = width
	r.height = height
	r.table.SetWidth(width - 4)
	r.table.SetHeight(height - 4)
}

// SetFocused sets the focus state
func (r *Results) SetFocused(focused bool) {
	r.focused = focused
	r.table.Focus()
}

// IsFocused returns if results is focused
func (r Results) IsFocused() bool {
	return r.focused
}

// SetData sets the query results data
func (r *Results) SetData(columns []string, rows []map[string]interface{}) {
	r.columns = columns
	r.rows = rows
	r.rowCount = len(rows)
	r.message = ""
	r.isError = false
	r.page = 0

	// Convert to table format
	r.updateTable()
}

// SetError sets an error message
func (r *Results) SetError(err error) {
	r.message = err.Error()
	r.isError = true
	r.columns = nil
	r.rows = nil
	r.rowCount = 0
}

// SetMessage sets an info message
func (r *Results) SetMessage(msg string) {
	r.message = msg
	r.isError = false
}

// Clear clears the results
func (r *Results) Clear() {
	r.columns = nil
	r.rows = nil
	r.rowCount = 0
	r.message = ""
	r.isError = false
}

// updateTable updates the internal table with current data
func (r *Results) updateTable() {
	if len(r.columns) == 0 {
		return
	}

	// Calculate column widths
	colWidth := (r.width - 4) / len(r.columns)
	if colWidth < 10 {
		colWidth = 10
	}
	if colWidth > 30 {
		colWidth = 30
	}

	// Create table columns
	cols := make([]table.Column, len(r.columns))
	for i, col := range r.columns {
		cols[i] = table.Column{
			Title: strings.ToUpper(col),
			Width: colWidth,
		}
	}

	// Create table rows (paginated)
	start := r.page * r.pageSize
	end := start + r.pageSize
	if end > len(r.rows) {
		end = len(r.rows)
	}

	tableRows := make([]table.Row, 0)
	for i := start; i < end; i++ {
		row := r.rows[i]
		tableRow := make(table.Row, len(r.columns))
		for j, col := range r.columns {
			val := row[col]
			tableRow[j] = formatValue(val, colWidth-2)
		}
		tableRows = append(tableRows, tableRow)
	}

	r.table.SetColumns(cols)
	r.table.SetRows(tableRows)
}

// formatValue formats a value for display
func formatValue(val interface{}, maxWidth int) string {
	var str string
	switch v := val.(type) {
	case nil:
		str = "NULL"
	case []byte:
		str = string(v)
	default:
		str = fmt.Sprintf("%v", v)
	}

	// Truncate if too long
	if len(str) > maxWidth {
		str = str[:maxWidth-3] + "..."
	}

	return str
}

// Update handles input for the results
func (r Results) Update(msg tea.Msg) (Results, tea.Cmd) {
	if !r.focused {
		return r, nil
	}

	var cmd tea.Cmd
	r.table, cmd = r.table.Update(msg)
	return r, cmd
}

// View renders the results
func (r Results) View() string {
	style := r.styles.Normal
	if r.focused {
		style = r.styles.Focused
	}

	var content strings.Builder

	// Title with view mode
	title := "RESULTS"
	if r.rowCount > 0 {
		modeStr := ""
		switch r.viewMode {
		case ViewTable:
			modeStr = "Table"
		case ViewChartBar:
			modeStr = "Bar Chart"
		case ViewChartLine:
			modeStr = "Line Chart"
		case ViewChartPie:
			modeStr = "Pie Chart"
		}
		title = fmt.Sprintf("RESULTS (%d rows) - %s", r.rowCount, modeStr)
	}
	content.WriteString(r.styles.Title.Render(title))
	content.WriteString("\n")

	// Show message or content
	if r.message != "" {
		if r.isError {
			content.WriteString(r.styles.Error.Render("Error: " + r.message))
		} else {
			content.WriteString(r.styles.Info.Render(r.message))
		}
	} else if len(r.rows) > 0 {
		switch r.viewMode {
		case ViewChartBar:
			content.WriteString(r.renderBarChart())
		case ViewChartLine:
			content.WriteString(r.renderLineChart())
		case ViewChartPie:
			content.WriteString(r.renderPieChart())
		default:
			content.WriteString(r.table.View())
			
			// Pagination info
			if r.rowCount > r.pageSize {
				totalPages := (r.rowCount + r.pageSize - 1) / r.pageSize
				pageInfo := fmt.Sprintf("\nPage %d/%d", r.page+1, totalPages)
				content.WriteString(r.styles.Info.Render(pageInfo))
			}
		}
	} else {
		content.WriteString(r.styles.Info.Render("No results. Run a query to see data here."))
	}

	return style.
		Width(r.width).
		Height(r.height).
		Render(content.String())
}

// SetViewMode sets the current view mode
func (r *Results) SetViewMode(mode ViewMode) {
	r.viewMode = mode
}

// GetViewMode returns the current view mode
func (r Results) GetViewMode() ViewMode {
	return r.viewMode
}

// renderLineChart renders a line chart using asciigraph
func (r Results) renderLineChart() string {
	data, label := r.extractNumericData()
	if len(data) == 0 {
		return "No numeric data found for chart"
	}
	
	graph := asciigraph.Plot(data, 
		asciigraph.Height(r.height-6),
		asciigraph.Width(r.width-10),
		asciigraph.Caption(label),
	)
	return graph
}

// renderBarChart renders a simple bar chart
func (r Results) renderBarChart() string {
	data, label := r.extractNumericData()
	if len(data) == 0 {
		return "No numeric data found for chart"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Bar Chart: %s\n\n", label))
	
	maxVal := 0.0
	for _, v := range data {
		if v > maxVal {
			maxVal = v
		}
	}
	
	maxBarWidth := r.width - 20
	if maxBarWidth < 10 {
		maxBarWidth = 10
	}
	
	// Limit rows for chart
	limit := r.height - 6
	if limit > len(data) {
		limit = len(data)
	}
	
	for i := 0; i < limit; i++ {
		val := data[i]
		barLen := int((val / maxVal) * float64(maxBarWidth))
		bar := strings.Repeat("█", barLen)
		b.WriteString(fmt.Sprintf("%3d │ %s %.2f\n", i+1, bar, val))
	}
	
	return b.String()
}

// renderPieChart renders a simple pie chart (hamburger style)
func (r Results) renderPieChart() string {
	data, label := r.extractNumericData()
	if len(data) == 0 {
		return "No numeric data found for chart"
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Pie Chart (Distribution): %s\n\n", label))
	
	total := 0.0
	for _, v := range data {
		total += v
	}
	
	// Limit slices
	limit := r.height - 6
	if limit > len(data) {
		limit = len(data)
	}
	
	chars := []string{"█", "▓", "▒", "░"}
	
	for i := 0; i < limit; i++ {
		val := data[i]
		percent := (val / total) * 100
		char := chars[i%len(chars)]
		b.WriteString(fmt.Sprintf("%s %.1f%% (%.2f)\n", char, percent, val))
	}
	
	return b.String()
}

// extractNumericData finds the first numeric column and returns data
func (r Results) extractNumericData() ([]float64, string) {
	if len(r.rows) == 0 || len(r.columns) == 0 {
		return nil, ""
	}

	// Find first numeric column
	targetCol := ""
	for _, col := range r.columns {
		// Check first row value
		val := r.rows[0][col]
		switch val.(type) {
		case int, int64, float64, float32:
			targetCol = col
			break
		}
		if targetCol != "" {
			break
		}
	}

	if targetCol == "" {
		return nil, ""
	}

	data := make([]float64, 0, len(r.rows))
	for _, row := range r.rows {
		val := row[targetCol]
		var floatVal float64
		switch v := val.(type) {
		case int:
			floatVal = float64(v)
		case int64:
			floatVal = float64(v)
		case float32:
			floatVal = float64(v)
		case float64:
			floatVal = v
		default:
			continue
		}
		data = append(data, floatVal)
	}

	return data, targetCol
}

// NextPage moves to the next page
func (r *Results) NextPage() {
	maxPage := (r.rowCount - 1) / r.pageSize
	if r.page < maxPage {
		r.page++
		r.updateTable()
	}
}

// PrevPage moves to the previous page
func (r *Results) PrevPage() {
	if r.page > 0 {
		r.page--
		r.updateTable()
	}
}

// GetRowCount returns the number of rows
func (r Results) GetRowCount() int {
	return r.rowCount
}

// CopySelectedRow copies the selected row data to clipboard
func (r Results) CopySelectedRow() error {
	cursor := r.table.Cursor()
	start := r.page * r.pageSize
	rowIdx := start + cursor
	
	if rowIdx >= len(r.rows) {
		return fmt.Errorf("no row selected")
	}
	
	row := r.rows[rowIdx]
	var values []string
	for _, col := range r.columns {
		val := row[col]
		values = append(values, fmt.Sprintf("%v", val))
	}
	
	text := strings.Join(values, "\t")
	return clipboard.WriteAll(text)
}

// CopyAllData copies all data to clipboard as TSV
func (r Results) CopyAllData() error {
	if len(r.rows) == 0 {
		return fmt.Errorf("no data to copy")
	}
	
	var lines []string
	
	// Header
	lines = append(lines, strings.Join(r.columns, "\t"))
	
	// Rows
	for _, row := range r.rows {
		var values []string
		for _, col := range r.columns {
			val := row[col]
			values = append(values, fmt.Sprintf("%v", val))
		}
		lines = append(lines, strings.Join(values, "\t"))
	}
	
	text := strings.Join(lines, "\n")
	return clipboard.WriteAll(text)
}
