// ABOUTME: The editor model handles text display, cursor, and viewport.
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
	docID      string
	lines      []string
	cursorRow  int
	cursorCol  int
	viewportTop int
	lockHolder string
	users      []string
	hasLock    bool
	showHelp   bool
	width      int
	height     int
}

// NewEditorModel creates an editor with sample content for now.
func NewEditorModel(docID string) EditorModel {
	return EditorModel{
		docID: docID,
		lines: []string{
			"# Welcome to Texty",
			"",
			"A collaborative terminal text editor.",
			"",
			"## Features",
			"- Real-time collaboration",
			"- Pessimistic locking",
			"- Multiple documents",
			"",
		},
		cursorRow:  0,
		cursorCol:  0,
		lockHolder: "tyler",
		users:      []string{"tyler", "jordan"},
		hasLock:    true,
	}
}

func (m EditorModel) Init() tea.Cmd {
	return nil
}

func (m EditorModel) Update(raw tea.Msg) (EditorModel, tea.Cmd) {
	switch v := raw.(type) {
	case tea.KeyMsg:
		switch v.String() {
		case "ctrl+q":
			return m, func() tea.Msg {
				return msg.SwitchScreen{Screen: msg.ScreenDocList}
			}
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = v.Width
		m.height = v.Height
	}
	return m, nil
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

	// Text area height = total height - header - status - 1 for border
	textHeight := m.height - 2
	if textHeight < 1 {
		textHeight = 1
	}

	// Border between header and text
	borderStyle := lipgloss.NewStyle().
		Foreground(ui.ColorBorder).
		Width(m.width)
	border := borderStyle.Render(strings.Repeat("─", m.width))

	// Render text lines with gutter
	totalLines := len(m.lines)
	gutterW := ui.GutterWidth(totalLines)
	textWidth := m.width - gutterW

	var textRows []string
	for i := 0; i < textHeight; i++ {
		lineIdx := m.viewportTop + i
		var row string
		if lineIdx < totalLines {
			gutter := ui.GutterLine(lineIdx+1, totalLines)
			line := m.lines[lineIdx]

			// Cursor rendering
			if lineIdx == m.cursorRow && m.hasLock {
				line = renderCursor(line, m.cursorCol)
			}

			// Truncate or pad line to fit
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
		textArea = ui.HelpOverlay(m.width, textHeight)
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
