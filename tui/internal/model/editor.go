// ABOUTME: The editor model handles text editing, cursor, and viewport.
// ABOUTME: This is the core editing surface for collaborative document editing.

package model

import (
	"context"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tysont/texty/tui/internal/api"
	"github.com/tysont/texty/tui/internal/msg"
	"github.com/tysont/texty/tui/internal/ui"
)

const (
	saveDebounce = 500 * time.Millisecond
	idleTimeout  = 3 * time.Second
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

	// Networking
	client    *api.Client
	userID    string
	username  string
	sseCtx    context.Context
	sseCancel context.CancelFunc
	isTyping  bool
	dirty     bool

	// Timers — generation counters to ignore stale ticks
	saveGen  int
	idleGen  int
}

// NewEditorModel creates an editor connected to the backend.
func NewEditorModel(docID, userID, username, serverURL string) EditorModel {
	ctx, cancel := context.WithCancel(context.Background())
	return EditorModel{
		docID:     docID,
		lines:     []string{""},
		cursorRow: 0,
		cursorCol: 0,
		hasLock:   false,
		client:    api.NewClient(serverURL),
		userID:    userID,
		username:  username,
		sseCtx:    ctx,
		sseCancel: cancel,
	}
}

func (m EditorModel) sseURL() string {
	return m.client.SubscribeURL(m.docID, m.userID, m.username)
}

func (m EditorModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchInitialText(),
		api.ListenSSE(m.sseCtx, m.sseURL()),
	)
}

func (m EditorModel) fetchInitialText() tea.Cmd {
	docID := m.docID
	return func() tea.Msg {
		state, err := m.client.GetText(docID)
		if err != nil {
			return msg.SSEError{Err: err}
		}
		return msg.SSEUpdate{
			Text:       state.Text,
			LockHolder: state.LockHolder,
		}
	}
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

// fullText joins all lines into a single string.
func (m EditorModel) fullText() string {
	return strings.Join(m.lines, "\n")
}

// setTextFromString replaces the editor content from a full text string.
func (m *EditorModel) setTextFromString(text string) {
	if text == "" {
		m.lines = []string{""}
	} else {
		m.lines = strings.Split(text, "\n")
	}
}

func (m EditorModel) Update(raw tea.Msg) (EditorModel, tea.Cmd) {
	switch v := raw.(type) {
	case tea.WindowSizeMsg:
		m.width = v.Width
		m.height = v.Height
		return m, nil

	case msg.SSEUpdate:
		return m.handleSSEUpdate(v)

	case msg.SSEError:
		// Reconnect after a delay
		return m, tea.Tick(2*time.Second, func(time.Time) tea.Msg {
			return sseReconnect{}
		})

	case sseReconnect:
		return m, api.ListenSSE(m.sseCtx, m.sseURL())

	case msg.SaveTick:
		if v.Gen == m.saveGen && m.dirty && m.hasLock {
			m.dirty = false
			return m, m.saveText()
		}
		return m, nil

	case msg.IdleTimeout:
		if v.Gen == m.idleGen && m.hasLock {
			m.isTyping = false
			cmds := []tea.Cmd{m.saveText(), m.releaseLock()}
			return m, tea.Batch(cmds...)
		}
		return m, nil

	case msg.TextSaved:
		return m, nil

	case msg.LockAcquired:
		m.hasLock = v.Success
		return m, nil

	case msg.LockReleased:
		m.hasLock = false
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(v)
	}
	return m, nil
}

type sseReconnect struct{}

func (m EditorModel) handleSSEUpdate(v msg.SSEUpdate) (EditorModel, tea.Cmd) {
	m.lockHolder = v.LockHolder
	if v.Users != nil {
		m.users = v.Users
	}

	// Don't overwrite local text while we're actively typing
	if !m.isTyping {
		m.setTextFromString(v.Text)
		m.clampCursor()
	}

	// Re-subscribe for the next SSE event
	return m, api.ListenSSE(m.sseCtx, m.sseURL())
}

func (m EditorModel) handleKeyMsg(v tea.KeyMsg) (EditorModel, tea.Cmd) {
	if m.showHelp {
		if v.String() == "?" || v.String() == "esc" {
			m.showHelp = false
		}
		return m, nil
	}

	switch v.String() {
	case "ctrl+c":
		m.sseCancel()
		if m.hasLock {
			_ = m.client.PostText(m.docID, m.userID, m.fullText())
			_ = m.client.ReleaseLock(m.docID, m.userID)
		}
		return m, tea.Quit

	case "ctrl+q":
		m.sseCancel()
		if m.hasLock {
			_ = m.client.PostText(m.docID, m.userID, m.fullText())
			_ = m.client.ReleaseLock(m.docID, m.userID)
		}
		return m, func() tea.Msg {
			return msg.SwitchScreen{Screen: msg.ScreenDocList}
		}

	case "?":
		m.showHelp = true
		return m, nil

	case "ctrl+s":
		if m.hasLock {
			m.dirty = false
			return m, m.saveText()
		}
		return m, nil

	// Cursor movement (no lock needed)
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

	// Editing keys — need lock
	case "enter":
		return m.editAction(func() { m.insertNewline() })
	case "backspace":
		return m.editAction(func() { m.deleteBack() })
	case "delete":
		return m.editAction(func() { m.deleteForward() })
	case "tab":
		return m.editAction(func() { m.insertText("    ") })

	default:
		if v.Type == tea.KeyRunes {
			return m.editAction(func() { m.insertText(string(v.Runes)) })
		}
	}

	return m, nil
}

// editAction acquires the lock if needed, performs the edit, and schedules save/idle timers.
func (m EditorModel) editAction(action func()) (EditorModel, tea.Cmd) {
	// If someone else holds the lock, ignore the edit
	if m.lockHolder != "" && m.lockHolder != m.userID && !m.hasLock {
		return m, nil
	}

	action()
	m.isTyping = true
	m.dirty = true
	m.scrollToCursor()

	var cmds []tea.Cmd

	// Acquire lock on first edit
	if !m.hasLock {
		cmds = append(cmds, m.acquireLock())
	}

	// Schedule debounced save (new generation invalidates previous tick)
	m.saveGen++
	saveGen := m.saveGen
	cmds = append(cmds, tea.Tick(saveDebounce, func(time.Time) tea.Msg {
		return msg.SaveTick{Gen: saveGen}
	}))

	// Schedule idle timeout (new generation invalidates previous tick)
	m.idleGen++
	idleGen := m.idleGen
	cmds = append(cmds, tea.Tick(idleTimeout, func(time.Time) tea.Msg {
		return msg.IdleTimeout{Gen: idleGen}
	}))

	return m, tea.Batch(cmds...)
}

func (m EditorModel) acquireLock() tea.Cmd {
	docID, userID := m.docID, m.userID
	return func() tea.Msg {
		success, _ := m.client.AcquireLock(docID, userID)
		return msg.LockAcquired{Success: success}
	}
}

func (m EditorModel) releaseLock() tea.Cmd {
	docID, userID := m.docID, m.userID
	return func() tea.Msg {
		_ = m.client.ReleaseLock(docID, userID)
		return msg.LockReleased{}
	}
}

func (m EditorModel) saveText() tea.Cmd {
	docID, userID, text := m.docID, m.userID, m.fullText()
	return func() tea.Msg {
		_ = m.client.PostText(docID, userID, text)
		return msg.TextSaved{}
	}
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
		m.lines[m.cursorRow] += m.lines[m.cursorRow+1]
		m.lines = append(m.lines[:m.cursorRow+1], m.lines[m.cursorRow+2:]...)
	}
}

func (m EditorModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	header := ui.HeaderBar(m.width,
		ui.HeaderAccent("texty"),
		ui.HeaderAccent(m.docID),
		ui.HeaderConnected(len(m.users)),
		ui.HeaderLock(m.lockHolder),
	)

	status := ui.StatusBar(m.width, m.cursorRow, m.cursorCol, m.hasLock)

	th := m.textHeight()

	borderStyle := lipgloss.NewStyle().Foreground(ui.ColorBorder)
	border := borderStyle.Render(strings.Repeat("─", m.width))

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
