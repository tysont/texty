// ABOUTME: Document list screen for browsing, creating, and deleting documents.
// ABOUTME: Fetches the document index from the backend and renders a selectable list.

package model

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tysont/texty/tui/internal/api"
	"github.com/tysont/texty/tui/internal/msg"
	"github.com/tysont/texty/tui/internal/ui"
)

// DocListModel is the document picker screen.
type DocListModel struct {
	docs     []api.DocSummary
	cursor   int
	client   *api.Client
	username string
	width    int
	height   int

	// Inline input for new document
	creating   bool
	newDocName string

	// Inline confirmation for delete
	deleting bool

	err string
}

// NewDocListModel creates a document list connected to the backend.
func NewDocListModel(serverURL, username string) DocListModel {
	return DocListModel{
		client:   api.NewClient(serverURL),
		username: username,
	}
}

func (m DocListModel) Init() tea.Cmd {
	return m.fetchDocs()
}

type docsListedMsg struct {
	docs []api.DocSummary
}
type docListError struct{ err string }

func (m DocListModel) fetchDocs() tea.Cmd {
	return func() tea.Msg {
		docs, err := m.client.ListDocs()
		if err != nil {
			return docListError{err: err.Error()}
		}
		return docsListedMsg{docs: docs}
	}
}

func (m DocListModel) Update(raw tea.Msg) (DocListModel, tea.Cmd) {
	switch v := raw.(type) {
	case tea.WindowSizeMsg:
		m.width = v.Width
		m.height = v.Height

	case docsListedMsg:
		m.docs = v.docs
		m.err = ""
		if m.cursor >= len(m.docs) {
			m.cursor = max(0, len(m.docs)-1)
		}

	case docListError:
		m.err = v.err

	case msg.DocCreated:
		return m, func() tea.Msg {
			return msg.SwitchScreen{Screen: msg.ScreenEditor, DocID: v.ID}
		}

	case msg.DocDeleted:
		m.deleting = false
		return m, m.fetchDocs()

	case tea.KeyMsg:
		return m.handleKey(v)
	}
	return m, nil
}

func (m DocListModel) handleKey(v tea.KeyMsg) (DocListModel, tea.Cmd) {
	if m.creating {
		switch v.String() {
		case "esc":
			m.creating = false
			m.newDocName = ""
		case "enter":
			name := strings.TrimSpace(m.newDocName)
			if name != "" {
				m.creating = false
				return m, m.createDoc(name)
			}
		case "backspace":
			if len(m.newDocName) > 0 {
				m.newDocName = m.newDocName[:len(m.newDocName)-1]
			}
		default:
			if v.Type == tea.KeyRunes {
				m.newDocName += string(v.Runes)
			}
		}
		return m, nil
	}

	if m.deleting {
		switch v.String() {
		case "y":
			if m.cursor < len(m.docs) {
				docID := m.docs[m.cursor].ID
				m.deleting = false
				return m, m.deleteDoc(docID)
			}
		default:
			m.deleting = false
		}
		return m, nil
	}

	switch v.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "j", "down":
		if m.cursor < len(m.docs)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "enter":
		if m.cursor < len(m.docs) {
			docID := m.docs[m.cursor].ID
			return m, func() tea.Msg {
				return msg.SwitchScreen{Screen: msg.ScreenEditor, DocID: docID}
			}
		}
	case "n":
		m.creating = true
		m.newDocName = ""
	case "d":
		if len(m.docs) > 0 {
			m.deleting = true
		}
	case "r":
		return m, m.fetchDocs()
	}
	return m, nil
}

func (m DocListModel) createDoc(name string) tea.Cmd {
	return func() tea.Msg {
		id, err := m.client.CreateDoc(name)
		if err != nil {
			return docListError{err: err.Error()}
		}
		return msg.DocCreated{ID: id}
	}
}

func (m DocListModel) deleteDoc(docID string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.DeleteDoc(docID)
		if err != nil {
			return docListError{err: err.Error()}
		}
		return msg.DocDeleted{ID: docID}
	}
}

func (m DocListModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	header := ui.HeaderBar(m.width,
		ui.HeaderAccent("texty"),
		lipgloss.NewStyle().Foreground(ui.ColorTextDim).Background(ui.ColorSurface).Render("documents"),
		lipgloss.NewStyle().Foreground(ui.ColorAccent).Background(ui.ColorSurface).Render("@"+m.username),
	)

	borderStyle := lipgloss.NewStyle().Foreground(ui.ColorBorder)
	border := borderStyle.Render(strings.Repeat("─", m.width))

	status := ui.DocListStatusBar(m.width)

	contentHeight := m.height - 3
	if contentHeight < 1 {
		contentHeight = 1
	}

	var lines []string

	lines = append(lines, "")

	titleStyle := lipgloss.NewStyle().Foreground(ui.ColorText).Bold(true).Padding(0, 2)
	lines = append(lines, titleStyle.Render("Documents"))
	lines = append(lines, "")

	if m.err != "" {
		errStyle := lipgloss.NewStyle().Foreground(ui.ColorError).Padding(0, 2)
		lines = append(lines, errStyle.Render("Error: "+m.err))
		lines = append(lines, "")
	}

	if len(m.docs) == 0 && m.err == "" {
		dimStyle := lipgloss.NewStyle().Foreground(ui.ColorTextDim).Padding(0, 2)
		lines = append(lines, dimStyle.Render("No documents yet. Press 'n' to create one."))
	}

	for i, doc := range m.docs {
		pointer := "  "
		nameStyle := lipgloss.NewStyle().Foreground(ui.ColorTextDim)
		if i == m.cursor {
			pointer = "> "
			nameStyle = lipgloss.NewStyle().Foreground(ui.ColorAccent).Bold(true)
		}

		meta := lipgloss.NewStyle().Foreground(ui.ColorTextDim).Render(
			fmt.Sprintf("%d connected    %d lines", doc.ConnectedUsers, doc.LineCount),
		)

		name := nameStyle.Render(doc.Name)
		nameWidth := 24
		padding := nameWidth - lipgloss.Width(doc.Name)
		if padding < 2 {
			padding = 2
		}
		gap := strings.Repeat(" ", padding)

		line := "   " + pointer + name + gap + meta
		lines = append(lines, line)
	}

	if m.creating {
		lines = append(lines, "")
		prompt := lipgloss.NewStyle().Foreground(ui.ColorAccent).Padding(0, 2).
			Render("New document name: " + m.newDocName + "█")
		lines = append(lines, prompt)
	}

	if m.deleting && m.cursor < len(m.docs) {
		lines = append(lines, "")
		prompt := lipgloss.NewStyle().Foreground(ui.ColorWarning).Padding(0, 2).
			Render(fmt.Sprintf("Delete '%s'? (y/n)", m.docs[m.cursor].Name))
		lines = append(lines, prompt)
	}

	for len(lines) < contentHeight {
		lines = append(lines, "")
	}
	if len(lines) > contentHeight {
		lines = lines[:contentHeight]
	}

	content := strings.Join(lines, "\n")
	return header + "\n" + border + "\n" + content + "\n" + status
}
