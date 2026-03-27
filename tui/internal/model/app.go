// ABOUTME: Top-level Bubble Tea model that manages screen transitions.
// ABOUTME: Delegates Init/Update/View to the currently active sub-model.

package model

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tysont/texty/tui/internal/config"
	"github.com/tysont/texty/tui/internal/msg"
)

// AppModel is the root model that switches between screens.
type AppModel struct {
	screen    msg.ScreenType
	usernameM UsernameModel
	docList   DocListModel
	editor    EditorModel
	serverURL string
	userID    string
	username  string
	width     int
	height    int
}

// NewAppModel creates the app, starting at the username screen if needed.
func NewAppModel(serverURL, userID string) AppModel {
	cfg := config.Load()

	m := AppModel{
		serverURL: serverURL,
		userID:    userID,
		username:  cfg.Username,
	}

	if cfg.Username == "" {
		m.screen = msg.ScreenUsername
		m.usernameM = NewUsernameModel()
	} else {
		m.screen = msg.ScreenDocList
		m.docList = NewDocListModel(serverURL, cfg.Username)
	}

	return m
}

func (m AppModel) Init() tea.Cmd {
	switch m.screen {
	case msg.ScreenDocList:
		return m.docList.Init()
	}
	return nil
}

func (m AppModel) Update(raw tea.Msg) (tea.Model, tea.Cmd) {
	switch v := raw.(type) {
	case tea.WindowSizeMsg:
		m.width = v.Width
		m.height = v.Height

	case msg.SwitchScreen:
		return m.switchScreen(v)

	case usernameSet:
		m.username = v.name
		return m.switchScreen(msg.SwitchScreen{Screen: msg.ScreenDocList})

	case msg.ForceQuit:
		return m, tea.Quit
	}

	switch m.screen {
	case msg.ScreenUsername:
		um, cmd := m.usernameM.Update(raw)
		m.usernameM = um
		return m, cmd
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
	case msg.ScreenUsername:
		return m.usernameM.View()
	case msg.ScreenDocList:
		return m.docList.View()
	case msg.ScreenEditor:
		return m.editor.View()
	default:
		return "Unknown screen"
	}
}
