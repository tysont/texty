// ABOUTME: The editor model handles text editing, cursor, and viewport.
// ABOUTME: This is the core editing surface for collaborative document editing.

package model

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tysont/texty/tui/internal/msg"
	"github.com/tysont/texty/tui/internal/ui"
)

// EditorModel is the main text editor screen.
type EditorModel struct {
	docID       string
	lines       []string
	cursorRow   int
	cursorCol   int
	viewportTop int
	lockHolder  string
	users       []string
	hasLock     bool
	showHelp    bool
	width       int
	height      int
}

// NewEditorModel creates an editor with an empty document.
func NewEditorModel(docID string) EditorModel {
	return EditorModel{
		docID:      docID,
		lines:      []string{""},
		cursorRow:  0,
		cursorCol:  0,
		lockHolder: "",
		users:      []string{},
		hasLock:    true, // local-only for now
	}
}

func (m EditorModel) Init() tea.Cmd {
	return nil
}

// textHeight returns the number of visible text lines.
func (m EditorModel) textHeight() int {
	h := m.height - 3 // header + border + status
	if h < 1 {
		h = 1
	}
	return h
}

// clampCursor ensures cursor stays within valid bounds.
func (m *EditorModel) clampCursor() {
	if m.cursorRow < 0 {
		m.cursorRow = 0
	}
	if m.cursorRow >= len(m.lines) {
		m.cursorRow = len(m.lines) - 1
	}
	lineLen := len([]rune(m.lines[m.cursorRow]))
	if m.cursorCol < 0 {
		m.cursorCol = 0
	}
	if m.cursorCol > lineLen {
		m.cursorCol = lineLen
	}
}

// scrollToCursor adjusts viewport so cursor is visible.
func (m *EditorModel) scrollToCursor() {
	th := m.textHeight()
	if m.cursorRow < m.viewportTop {
		m.viewportTop = m.cursorRow
	}
	if m.cursorRow >= m.viewportTop+th {
		m.viewportTop = m.cursorRow - th + 1
	}
}

func (m EditorModel) Update(raw tea.Msg) (EditorModel, tea.Cmd) {
	switch v := raw.(type) {
	case tea.WindowSizeMsg:
		m.width = v.Width
		m.height = v.Height
		return m, nil

	case tea.KeyMsg:
		if m.showHelp {
			if v.String() == "?" || v.String() == "esc" {
				m.showHelp = false
			}
			return m, nil
		}

		switch v.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+q":
			return m, func() tea.Msg {
				return msg.SwitchScreen{Screen: msg.ScreenDocList}
			}
		case "?":
			m.showHelp = true
			return m, nil

		// Cursor movement
		case "up":
			m.cursorRow--
			m.clampCursor()
			m.scrollToCursor()
		case "down":
			m.cursorRow++
			m.clampCursor()
			m.scrollToCursor()
		case "left":
			if m.cursorCol > 0 {
				m.cursorCol--
			} else if m.cursorRow > 0 {
				m.cursorRow--
				m.cursorCol = len([]rune(m.lines[m.cursorRow]))
			}
			m.scrollToCursor()
		case "right":
			lineLen := len([]rune(m.lines[m.cursorRow]))
			if m.cursorCol < lineLen {
				m.cursorCol++
			} else if m.cursorRow < len(m.lines)-1 {
				m.cursorRow++
				m.cursorCol = 0
			}
			m.scrollToCursor()
		case "home", "ctrl+a":
			m.cursorCol = 0
		case "end", "ctrl+e":
			m.cursorCol = len([]rune(m.lines[m.cursorRow]))

		// Editing
		case "enter":
			m.insertNewline()
			m.scrollToCursor()
		case "backspace":
			m.deleteBack()
			m.scrollToCursor()
		case "delete":
			m.deleteForward()
		case "tab":
			m.insertText("    ")

		default:
			// Insert printable characters
			if v.Type == tea.KeyRunes {
				m.insertText(string(v.Runes))
			}
		}
	}
	return m, nil
}

// insertText inserts a string at the cursor position.
func (m *EditorModel) insertText(s string) {
	runes := []rune(m.lines[m.cursorRow])
	col := m.cursorCol
	if col > len(runes) {
		col = len(runes)
	}
	insertRunes := []rune(s)
	newRunes := make([]rune, 0, len(runes)+len(insertRunes))
	newRunes = append(newRunes, runes[:col]...)
	newRunes = append(newRunes, insertRunes...)
	newRunes = append(newRunes, runes[col:]...)
	m.lines[m.cursorRow] = string(newRunes)
	m.cursorCol = col + len(insertRunes)
}

// insertNewline splits the current line at the cursor.
func (m *EditorModel) insertNewline() {
	runes := []rune(m.lines[m.cursorRow])
	col := m.cursorCol
	if col > len(runes) {
		col = len(runes)
	}

	before := string(runes[:col])
	after := string(runes[col:])

	m.lines[m.cursorRow] = before

	// Insert new line after current
	newLines := make([]string, 0, len(m.lines)+1)
	newLines = append(newLines, m.lines[:m.cursorRow+1]...)
	newLines = append(newLines, after)
	newLines = append(newLines, m.lines[m.cursorRow+1:]...)
	m.lines = newLines

	m.cursorRow++
	m.cursorCol = 0
}

// deleteBack removes the character before the cursor (backspace).
func (m *EditorModel) deleteBack() {
	if m.cursorCol > 0 {
		runes := []rune(m.lines[m.cursorRow])
		col := m.cursorCol
		if col > len(runes) {
			col = len(runes)
		}
		newRunes := make([]rune, 0, len(runes)-1)
		newRunes = append(newRunes, runes[:col-1]...)
		newRunes = append(newRunes, runes[col:]...)
		m.lines[m.cursorRow] = string(newRunes)
		m.cursorCol = col - 1
	} else if m.cursorRow > 0 {
		// Merge with previous line
		prevLen := len([]rune(m.lines[m.cursorRow-1]))
		m.lines[m.cursorRow-1] += m.lines[m.cursorRow]
		m.lines = append(m.lines[:m.cursorRow], m.lines[m.cursorRow+1:]...)
		m.cursorRow--
		m.cursorCol = prevLen
	}
}

// deleteForward removes the character at the cursor (delete key).
func (m *EditorModel) deleteForward() {
	runes := []rune(m.lines[m.cursorRow])
	if m.cursorCol < len(runes) {
		newRunes := make([]rune, 0, len(runes)-1)
		newRunes = append(newRunes, runes[:m.cursorCol]...)
		newRunes = append(newRunes, runes[m.cursorCol+1:]...)
		m.lines[m.cursorRow] = string(newRunes)
	} else if m.cursorRow < len(m.lines)-1 {
		// Merge with next line
		m.lines[m.cursorRow] += m.lines[m.cursorRow+1]
		m.lines = append(m.lines[:m.cursorRow+1], m.lines[m.cursorRow+2:]...)
	}
}

func (m EditorModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	// Header
	header := ui.HeaderBar(m.width,
		ui.HeaderAccent("texty"),
		ui.HeaderAccent(m.docID),
		ui.HeaderConnected(len(m.users)),
		ui.HeaderLock(m.lockHolder),
	)

	// Status bar
	status := ui.StatusBar(m.width, m.cursorRow, m.cursorCol, m.hasLock)

	th := m.textHeight()

	// Border between header and text
	borderStyle := lipgloss.NewStyle().Foreground(ui.ColorBorder)
	border := borderStyle.Render(strings.Repeat("─", m.width))

	// Render text lines with gutter
	totalLines := len(m.lines)
	gutterW := ui.GutterWidth(totalLines)
	textWidth := m.width - gutterW
	if textWidth < 1 {
		textWidth = 1
	}

	var textRows []string
	for i := 0; i < th; i++ {
		lineIdx := m.viewportTop + i
		var row string
		if lineIdx < totalLines {
			gutter := ui.GutterLine(lineIdx+1, totalLines)
			line := m.lines[lineIdx]

			// Cursor rendering
			if lineIdx == m.cursorRow {
				line = renderCursor(line, m.cursorCol)
			}

			lineStyle := lipgloss.NewStyle().
				Foreground(ui.ColorText).
				Width(textWidth)
			row = gutter + lineStyle.Render(line)
		} else {
			row = ui.GutterEmpty(totalLines) +
				lipgloss.NewStyle().Width(textWidth).Render("")
		}
		textRows = append(textRows, row)
	}

	textArea := strings.Join(textRows, "\n")

	if m.showHelp {
		textArea = ui.HelpOverlay(m.width, th)
	}

	return header + "\n" + border + "\n" + textArea + "\n" + status
}

// renderCursor inserts a block cursor at the given column position.
func renderCursor(line string, col int) string {
	runes := []rune(line)
	if col > len(runes) {
		col = len(runes)
	}

	var ch string
	if col < len(runes) {
		ch = string(runes[col])
	} else {
		ch = " "
	}

	cursor := lipgloss.NewStyle().
		Reverse(true).
		Render(ch)

	var after string
	if col+1 < len(runes) {
		after = string(runes[col+1:])
	}
	return string(runes[:col]) + cursor + after
}
