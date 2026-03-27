// ABOUTME: Top-level Bubble Tea model that manages screen transitions.
// ABOUTME: Delegates Init/Update/View to the currently active sub-model.

package model

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tysont/texty/tui/internal/msg"
)

// AppModel is the root model that switches between screens.
type AppModel struct {
	screen    msg.ScreenType
	docList   DocListModel
	editor    EditorModel
	serverURL string
	userID    string
	username  string
	width     int
	height    int
}

// NewAppModel creates the app pointing at the given server.
func NewAppModel(serverURL, userID, username string) AppModel {
	return AppModel{
		screen:    msg.ScreenDocList,
		docList:   NewDocListModel(serverURL, username),
		serverURL: serverURL,
		userID:    userID,
		username:  username,
	}
}

func (m AppModel) Init() tea.Cmd {
	return m.docList.Init()
}

func (m AppModel) Update(raw tea.Msg) (tea.Model, tea.Cmd) {
	switch v := raw.(type) {
	case tea.WindowSizeMsg:
		m.width = v.Width
		m.height = v.Height

	case msg.SwitchScreen:
		return m.switchScreen(v)

	case msg.ForceQuit:
		return m, tea.Quit
	}

	switch m.screen {
	case msg.ScreenDocList:
		docList, cmd := m.docList.Update(raw)
		m.docList = docList
		return m, cmd
	case msg.ScreenEditor:
		editor, cmd := m.editor.Update(raw)
		m.editor = editor
		return m, cmd
	}

	return m, nil
}

func (m AppModel) switchScreen(v msg.SwitchScreen) (tea.Model, tea.Cmd) {
	m.screen = v.Screen

	switch v.Screen {
	case msg.ScreenDocList:
		// Cancel SSE and clean up editor
		if m.editor.sseCancel != nil {
			m.editor.sseCancel()
		}
		m.docList = NewDocListModel(m.serverURL, m.username)
		m.docList.width = m.width
		m.docList.height = m.height
		return m, m.docList.Init()

	case msg.ScreenEditor:
		m.editor = NewEditorModel(v.DocID, m.userID, m.username, m.serverURL)
		m.editor.width = m.width
		m.editor.height = m.height
		return m, m.editor.Init()
	}

	return m, nil
}

func (m AppModel) View() string {
	switch m.screen {
	case msg.ScreenDocList:
		return m.docList.View()
	case msg.ScreenEditor:
		return m.editor.View()
	default:
		return "Unknown screen"
	}
}
