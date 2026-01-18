package terminal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// TableInfo represents a database table with row count (matching database.TableInfo)
type TableInfo struct {
	Name     string
	RowCount int
}

// TableRecords represents paginated records from a table
type TableRecords struct {
	Table      string
	TotalCount int
	Offset     int
	Limit      int
	Records    interface{}
}

// RecordFetcher is a function type for fetching table records
type RecordFetcher func(ctx context.Context, tableName string, offset, limit int, filters map[string]string, search string) (*TableRecords, error)

// ColumnFetcher is a function type for getting table columns
type ColumnFetcher func(tableName string) []string

// AllColumnsFetcher returns ALL columns for a table (for column selection UI)
type AllColumnsFetcher func(tableName string) []string

// dbView represents the current view state
type dbView int

const (
	viewTableList dbView = iota
	viewColumnSelect
	viewTableDetail
	viewRecordDetail
)

// Styles - Green theme
var (
	dbTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("34")). // Green
			MarginBottom(1)

	dbSubtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	dbHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	dbBaseStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	// Selection prefix symbol
	dbSelectPrefix = "▶ "
)

// tableItem implements list.Item for the table list
type tableItem struct {
	name     string
	rowCount int
}

func (i tableItem) Title() string       { return i.name }
func (i tableItem) Description() string { return fmt.Sprintf("%d records", i.rowCount) }
func (i tableItem) FilterValue() string { return i.name }

// prefixDelegate is a custom list delegate that shows a prefix for selected items
type prefixDelegate struct {
	normalStyle   lipgloss.Style
	selectedStyle lipgloss.Style
	descStyle     lipgloss.Style
	prefix        string
	spacing       int
}

func newPrefixDelegate() prefixDelegate {
	return prefixDelegate{
		normalStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		selectedStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("34")), // Green
		descStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		prefix:        dbSelectPrefix,
		spacing:       1,
	}
}

func (d prefixDelegate) Height() int                             { return 2 }
func (d prefixDelegate) Spacing() int                            { return d.spacing }
func (d prefixDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d prefixDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(tableItem)
	if !ok {
		return
	}

	// Determine if this item is selected
	isSelected := index == m.Index()

	var title, desc string
	if isSelected {
		title = d.selectedStyle.Render(d.prefix + i.Title())
		desc = d.descStyle.Render("  " + i.Description())
	} else {
		title = d.normalStyle.Render("  " + i.Title())
		desc = d.descStyle.Render("  " + i.Description())
	}

	_, _ = fmt.Fprintf(w, "%s\n%s", title, desc)
}

// DBTUI is the public interface for the database TUI
type DBTUI struct {
	program           *tea.Program
	tables            []TableInfo
	recordFetcher     RecordFetcher
	columnFetcher     ColumnFetcher
	allColumnsFetcher AllColumnsFetcher
	pageSize          int
}

// dbTUIModel is the internal tea.Model implementation
type dbTUIModel struct {
	// Current view state
	view dbView

	// Table list (main view)
	tableList list.Model
	tables    []TableInfo

	// Column selection view
	allColumns      []string        // All available columns for selected table
	selectedColumns map[string]bool // Toggle state for each column
	columnCursor    int             // Current cursor position in column list
	columnScroll    int             // Scroll offset for column list

	// Table detail view
	recordTable  table.Model
	paginator    paginator.Model
	searchInput  textinput.Model
	searchActive bool
	searchQuery  string

	// Record detail view
	recordViewport viewport.Model
	selectedRecord map[string]interface{}

	// Current selection
	selectedTable string
	columns       []string
	records       []map[string]interface{}
	totalRecords  int
	currentPage   int
	pageSize      int

	// Data fetchers
	recordFetcher     RecordFetcher
	columnFetcher     ColumnFetcher
	allColumnsFetcher AllColumnsFetcher

	// UI state
	width, height int
	ready         bool
	err           error
	loading       bool
}

// NewDBTUI creates a new database TUI
func NewDBTUI(tables []TableInfo, recordFetcher RecordFetcher, columnFetcher ColumnFetcher, allColumnsFetcher AllColumnsFetcher, pageSize int) *DBTUI {
	if pageSize <= 0 {
		pageSize = 20 // default
	}
	return &DBTUI{
		tables:            tables,
		recordFetcher:     recordFetcher,
		columnFetcher:     columnFetcher,
		allColumnsFetcher: allColumnsFetcher,
		pageSize:          pageSize,
	}
}

// Run starts the TUI (blocking)
func (t *DBTUI) Run() error {
	model := t.newModel()
	t.program = tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := t.program.Run()
	return err
}

func (t *DBTUI) newModel() dbTUIModel {
	// Create list items from tables
	items := make([]list.Item, len(t.tables))
	for i, tbl := range t.tables {
		items[i] = tableItem{name: tbl.Name, rowCount: tbl.RowCount}
	}

	// Create list model with custom prefix delegate
	delegate := newPrefixDelegate()

	tableList := list.New(items, delegate, 0, 0)
	tableList.Title = "Database Tables"
	tableList.SetShowStatusBar(false)
	tableList.SetFilteringEnabled(true)
	tableList.Styles.Title = dbTitleStyle
	tableList.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(lipgloss.Color("34")) // Green
	tableList.Styles.FilterCursor = lipgloss.NewStyle().Foreground(lipgloss.Color("34")) // Green

	// Create search input
	searchInput := textinput.New()
	searchInput.Placeholder = "Search records..."
	searchInput.CharLimit = 100
	searchInput.Width = 30

	// Create paginator with green dots
	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = t.pageSize
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("34")).Render("•") // Green
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("•")

	return dbTUIModel{
		view:              viewTableList,
		tableList:         tableList,
		tables:            t.tables,
		searchInput:       searchInput,
		paginator:         p,
		pageSize:          t.pageSize,
		recordFetcher:     t.recordFetcher,
		columnFetcher:     t.columnFetcher,
		allColumnsFetcher: t.allColumnsFetcher,
		selectedColumns:   make(map[string]bool),
	}
}

// Messages
type recordsLoadedMsg struct {
	records      []map[string]interface{}
	totalRecords int
	columns      []string
	err          error
}

func (m dbTUIModel) Init() tea.Cmd {
	return nil
}

func (m dbTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle global keys
		switch msg.String() {
		case "ctrl+c", "q":
			if m.view == viewTableList {
				return m, tea.Quit
			}
			// In other views, q goes back
			if m.view == viewRecordDetail {
				m.view = viewTableDetail
				return m, nil
			}
			if m.view == viewTableDetail && !m.searchActive {
				m.view = viewTableList
				m.selectedTable = ""
				m.records = nil
				return m, nil
			}
		case "esc":
			if m.searchActive {
				m.searchActive = false
				m.searchInput.Blur()
				return m, nil
			}
			if m.view == viewRecordDetail {
				m.view = viewTableDetail
				return m, nil
			}
			if m.view == viewTableDetail {
				m.view = viewColumnSelect
				return m, nil
			}
			if m.view == viewColumnSelect {
				m.view = viewTableList
				m.selectedTable = ""
				m.allColumns = nil
				m.selectedColumns = make(map[string]bool)
				m.columnCursor = 0
				return m, nil
			}
		}

		// View-specific key handling
		switch m.view {
		case viewTableList:
			return m.updateTableList(msg)
		case viewColumnSelect:
			return m.updateColumnSelect(msg)
		case viewTableDetail:
			return m.updateTableDetail(msg)
		case viewRecordDetail:
			return m.updateRecordDetail(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Update component sizes (accounting for 2-char left padding)
		m.tableList.SetSize(msg.Width-6, msg.Height-4)

		if m.recordViewport.Width > 0 {
			m.recordViewport.Width = msg.Width - 6
			m.recordViewport.Height = msg.Height - 8
		}

		return m, nil

	case recordsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.records = msg.records
		m.totalRecords = msg.totalRecords
		m.columns = msg.columns

		// Setup table with records
		m.setupRecordTable()

		// Setup paginator
		totalPages := (msg.totalRecords + m.pageSize - 1) / m.pageSize
		if totalPages == 0 {
			totalPages = 1
		}
		m.paginator.SetTotalPages(totalPages)

		m.view = viewTableDetail
		return m, nil
	}

	// Update current view component
	switch m.view {
	case viewTableList:
		m.tableList, cmd = m.tableList.Update(msg)
		cmds = append(cmds, cmd)
	case viewTableDetail:
		if m.searchActive {
			m.searchInput, cmd = m.searchInput.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			m.recordTable, cmd = m.recordTable.Update(msg)
			cmds = append(cmds, cmd)
		}
	case viewRecordDetail:
		m.recordViewport, cmd = m.recordViewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m dbTUIModel) updateTableList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		item, ok := m.tableList.SelectedItem().(tableItem)
		if ok {
			m.selectedTable = item.name
			m.currentPage = 0
			m.columnCursor = 0
			m.columnScroll = 0

			// Get all columns for the table
			if m.allColumnsFetcher != nil {
				m.allColumns = m.allColumnsFetcher(m.selectedTable)
			}

			// Initialize default selected columns
			m.selectedColumns = getDefaultSelectedColumns(m.selectedTable, m.allColumns)

			m.view = viewColumnSelect
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.tableList, cmd = m.tableList.Update(msg)
	return m, cmd
}

// assetsDefaultColumns are the columns selected by default for assets table
var assetsDefaultColumns = map[string]bool{
	"host": true, "host_ip": true, "title": true, "status_code": true,
	"words": true, "technologies": true, "labels": true, "source": true,
}

// getDefaultSelectedColumns returns default column selection for a table
// Assets: only specific columns selected by default
// Other tables: exclude id, created_at, updated_at
func getDefaultSelectedColumns(tableName string, allCols []string) map[string]bool {
	metadataCols := map[string]bool{"id": true, "created_at": true, "updated_at": true}
	selected := make(map[string]bool)

	for _, col := range allCols {
		if tableName == "assets" {
			// Assets: only select specific columns by default
			selected[col] = assetsDefaultColumns[col]
		} else {
			// Others: exclude metadata columns
			selected[col] = !metadataCols[col]
		}
	}
	return selected
}

// updateColumnSelect handles key events in the column selection view
func (m dbTUIModel) updateColumnSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.columnCursor > 0 {
			m.columnCursor--
			// Scroll up if needed
			if m.columnCursor < m.columnScroll {
				m.columnScroll = m.columnCursor
			}
		}
	case "down", "j":
		if m.columnCursor < len(m.allColumns)-1 {
			m.columnCursor++
			// Scroll down if needed (show ~15 items)
			visibleRows := m.height - 12
			if visibleRows < 5 {
				visibleRows = 5
			}
			if m.columnCursor >= m.columnScroll+visibleRows {
				m.columnScroll = m.columnCursor - visibleRows + 1
			}
		}
	case " ":
		// Toggle current column
		if m.columnCursor < len(m.allColumns) {
			col := m.allColumns[m.columnCursor]
			m.selectedColumns[col] = !m.selectedColumns[col]
		}
	case "a":
		// Select all
		for _, col := range m.allColumns {
			m.selectedColumns[col] = true
		}
	case "n":
		// Select none
		for _, col := range m.allColumns {
			m.selectedColumns[col] = false
		}
	case "enter":
		// Build columns list from selected
		m.columns = nil
		for _, col := range m.allColumns {
			if m.selectedColumns[col] {
				m.columns = append(m.columns, col)
			}
		}
		if len(m.columns) == 0 {
			// If no columns selected, use defaults
			m.columns = m.allColumns
		}
		m.loading = true
		return m, m.loadRecords()
	}
	return m, nil
}

func (m dbTUIModel) updateTableDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "/":
		if !m.searchActive {
			m.searchActive = true
			m.searchInput.Focus()
			return m, textinput.Blink
		}
	case "enter":
		if m.searchActive {
			m.searchActive = false
			m.searchInput.Blur()
			m.searchQuery = m.searchInput.Value()
			m.currentPage = 0
			m.loading = true
			return m, m.loadRecords()
		}
		// View record detail
		if len(m.records) > 0 {
			cursor := m.recordTable.Cursor()
			if cursor >= 0 && cursor < len(m.records) {
				m.selectedRecord = m.records[cursor]
				m.view = viewRecordDetail
				m.setupRecordViewport()
				return m, nil
			}
		}
	case "n":
		if !m.searchActive && m.paginator.Page < m.paginator.TotalPages-1 {
			m.paginator.NextPage()
			m.currentPage = m.paginator.Page
			m.loading = true
			return m, m.loadRecords()
		}
	case "p":
		if !m.searchActive && m.paginator.Page > 0 {
			m.paginator.PrevPage()
			m.currentPage = m.paginator.Page
			m.loading = true
			return m, m.loadRecords()
		}
	case "c":
		// Clear search
		if !m.searchActive {
			m.searchQuery = ""
			m.searchInput.SetValue("")
			m.currentPage = 0
			m.loading = true
			return m, m.loadRecords()
		}
	}

	var cmd tea.Cmd
	if m.searchActive {
		m.searchInput, cmd = m.searchInput.Update(msg)
	} else {
		m.recordTable, cmd = m.recordTable.Update(msg)
	}
	return m, cmd
}

func (m dbTUIModel) updateRecordDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.recordViewport, cmd = m.recordViewport.Update(msg)
	return m, cmd
}

func (m *dbTUIModel) loadRecords() tea.Cmd {
	// Capture the columns from model (they're set from column selection)
	selectedCols := m.columns

	return func() tea.Msg {
		if m.recordFetcher == nil {
			return recordsLoadedMsg{err: fmt.Errorf("no record fetcher configured")}
		}

		offset := m.currentPage * m.pageSize
		ctx := context.Background()

		result, err := m.recordFetcher(ctx, m.selectedTable, offset, m.pageSize, nil, m.searchQuery)
		if err != nil {
			return recordsLoadedMsg{err: err}
		}

		// Convert records to []map[string]interface{}
		jsonBytes, _ := json.Marshal(result.Records)
		var records []map[string]interface{}
		_ = json.Unmarshal(jsonBytes, &records)

		// Use selected columns from column selection view
		columns := selectedCols

		return recordsLoadedMsg{
			records:      records,
			totalRecords: result.TotalCount,
			columns:      columns,
		}
	}
}

func (m *dbTUIModel) setupRecordTable() {
	if len(m.columns) == 0 || len(m.records) == 0 {
		return
	}

	// Determine visible columns (max 6 for display)
	visibleCols := m.columns
	if len(visibleCols) > 6 {
		visibleCols = visibleCols[:6]
	}

	// Calculate column widths (accounting for 2-char left padding)
	colWidth := (m.width - 12) / len(visibleCols)
	if colWidth < 10 {
		colWidth = 10
	}
	if colWidth > 30 {
		colWidth = 30
	}

	// Create table columns
	columns := make([]table.Column, len(visibleCols))
	for i, col := range visibleCols {
		columns[i] = table.Column{
			Title: strings.ToUpper(col),
			Width: colWidth,
		}
	}

	// Create table rows
	rows := make([]table.Row, len(m.records))
	for i, record := range m.records {
		row := make(table.Row, len(visibleCols))
		for j, col := range visibleCols {
			row[j] = formatCellValue(record[col], colWidth-3)
		}
		rows[i] = row
	}

	// Create table
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(m.height-12),
	)

	// Style the table - green theme with prefix
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("34")). // Green
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("34")) // Green
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("34")). // Green
		Background(lipgloss.NoColor{}).   // No background
		Bold(false)
	t.SetStyles(s)

	m.recordTable = t
}

func (m *dbTUIModel) setupRecordViewport() {
	// Format record as pretty JSON with syntax highlighting
	jsonBytes, _ := json.MarshalIndent(m.selectedRecord, "", "  ")

	// Wrap in markdown code block for syntax highlighting
	markdown := "```json\n" + string(jsonBytes) + "\n```"

	// Render with glamour
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(m.width-10),
	)

	var content string
	if err == nil {
		rendered, renderErr := renderer.Render(markdown)
		if renderErr == nil {
			content = rendered
		} else {
			content = string(jsonBytes) // fallback to plain JSON
		}
	} else {
		content = string(jsonBytes) // fallback to plain JSON
	}

	m.recordViewport = viewport.New(m.width-6, m.height-8)
	m.recordViewport.SetContent(content)
}

func formatCellValue(v interface{}, maxLen int) string {
	if v == nil {
		return ""
	}

	var s string
	switch val := v.(type) {
	case string:
		s = val
	case []interface{}, map[string]interface{}:
		b, _ := json.Marshal(val)
		s = string(b)
	default:
		s = fmt.Sprintf("%v", val)
	}

	// Truncate if too long
	if len(s) > maxLen {
		if maxLen > 3 {
			return s[:maxLen-3] + "..."
		}
		return s[:maxLen]
	}
	return s
}

func (m dbTUIModel) View() string {
	if !m.ready {
		return dbBaseStyle.Render("Loading...")
	}

	var content string
	switch m.view {
	case viewTableList:
		content = m.viewTableListView()
	case viewColumnSelect:
		content = m.viewColumnSelectView()
	case viewTableDetail:
		content = m.viewTableDetailView()
	case viewRecordDetail:
		content = m.viewRecordDetailView()
	}

	return dbBaseStyle.Render(content)
}

func (m dbTUIModel) viewTableListView() string {
	return m.tableList.View()
}

func (m dbTUIModel) viewColumnSelectView() string {
	var b strings.Builder

	// Count selected columns
	selectedCount := 0
	for _, col := range m.allColumns {
		if m.selectedColumns[col] {
			selectedCount++
		}
	}

	// Title and count
	title := dbTitleStyle.Render(fmt.Sprintf("Table: %s", m.selectedTable))
	countInfo := dbSubtitleStyle.Render(fmt.Sprintf("Selected: %d/%d columns", selectedCount, len(m.allColumns)))
	b.WriteString(title + "    " + countInfo + "\n\n")

	b.WriteString(dbSubtitleStyle.Render("Select columns to display:") + "\n\n")

	// Calculate visible rows
	visibleRows := m.height - 12
	if visibleRows < 5 {
		visibleRows = 5
	}

	// Column list with checkboxes - green theme with prefix selection
	checkboxStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("34")) // Green
	uncheckStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("34")) // Green

	endIdx := m.columnScroll + visibleRows
	if endIdx > len(m.allColumns) {
		endIdx = len(m.allColumns)
	}

	for i := m.columnScroll; i < endIdx; i++ {
		col := m.allColumns[i]
		var checkbox string
		if m.selectedColumns[col] {
			checkbox = checkboxStyle.Render("[x]")
		} else {
			checkbox = uncheckStyle.Render("[ ]")
		}

		if i == m.columnCursor {
			// Show prefix for selected row
			line := cursorStyle.Render(fmt.Sprintf("%s%s %s", dbSelectPrefix, checkbox, col))
			b.WriteString(line + "\n")
		} else {
			line := fmt.Sprintf("  %s %s", checkbox, col)
			b.WriteString(line + "\n")
		}
	}

	// Scroll indicator
	if len(m.allColumns) > visibleRows {
		scrollInfo := fmt.Sprintf("\n(%d-%d of %d)", m.columnScroll+1, endIdx, len(m.allColumns))
		b.WriteString(dbSubtitleStyle.Render(scrollInfo))
	}

	b.WriteString("\n\n")

	// Help
	help := "↑/↓: Navigate  Space: Toggle  a: All  n: None  Enter: View Records  Esc: Back"
	b.WriteString(dbHelpStyle.Render(help))

	return b.String()
}

func (m dbTUIModel) viewTableDetailView() string {
	var b strings.Builder

	// Title
	title := dbTitleStyle.Render(fmt.Sprintf("Table: %s", m.selectedTable))
	pageInfo := dbSubtitleStyle.Render(fmt.Sprintf("Page %d/%d  |  %d records", m.paginator.Page+1, m.paginator.TotalPages, m.totalRecords))
	b.WriteString(title + "  " + pageInfo + "\n\n")

	// Search bar
	searchLabel := "Search: "
	if m.searchActive {
		searchLabel = "Search: "
	}
	if m.searchQuery != "" {
		searchLabel = fmt.Sprintf("Search: \"%s\"  ", m.searchQuery)
	}
	b.WriteString(dbSubtitleStyle.Render(searchLabel))
	if m.searchActive {
		b.WriteString(m.searchInput.View())
	}
	b.WriteString("\n\n")

	// Loading indicator
	if m.loading {
		b.WriteString("Loading records...\n")
		return b.String()
	}

	// Error
	if m.err != nil {
		b.WriteString(fmt.Sprintf("Error: %v\n", m.err))
		return b.String()
	}

	// Table
	if len(m.records) == 0 {
		b.WriteString("No records found.\n")
	} else {
		b.WriteString(m.recordTable.View())
	}

	b.WriteString("\n")

	// Pagination
	b.WriteString(m.paginator.View())
	b.WriteString("\n\n")

	// Help
	help := "↑/↓: Navigate  Enter: View Detail  /: Search  n/p: Page  c: Clear Search  Esc: Back  q: Quit"
	b.WriteString(dbHelpStyle.Render(help))

	return b.String()
}

func (m dbTUIModel) viewRecordDetailView() string {
	var b strings.Builder

	title := dbTitleStyle.Render("Record Detail")
	b.WriteString(title + "\n\n")

	b.WriteString(m.recordViewport.View())
	b.WriteString("\n\n")

	help := "↑/↓: Scroll  Esc: Back  q: Quit"
	b.WriteString(dbHelpStyle.Render(help))

	return b.String()
}
