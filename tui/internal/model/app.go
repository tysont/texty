// ABOUTME: Top-level Bubble Tea model that manages screen transitions.
// ABOUTME: Delegates Init/Update/View to the currently active sub-model.

package model

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tysont/texty/tui/internal/msg"
)

// AppModel is the root model that switches between screens.
type AppModel struct {
	screen   msg.ScreenType
	editor   EditorModel
	width    int
	height   int
}

// NewAppModel creates the app starting at the editor screen with sample data.
func NewAppModel() AppModel {
	return AppModel{
		screen: msg.ScreenEditor,
		editor: NewEditorModel("meeting-notes"),
	}
}

func (m AppModel) Init() tea.Cmd {
	return nil
}

func (m AppModel) Update(raw tea.Msg) (tea.Model, tea.Cmd) {
	switch v := raw.(type) {
	case tea.WindowSizeMsg:
		m.width = v.Width
		m.height = v.Height

	case tea.KeyMsg:
		if v.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case msg.SwitchScreen:
		m.screen = v.Screen
		return m, nil

	case msg.ForceQuit:
		return m, tea.Quit
	}

	switch m.screen {
	case msg.ScreenEditor:
		editor, cmd := m.editor.Update(raw)
		m.editor = editor
		return m, cmd
	}

	return m, nil
}

func (m AppModel) View() string {
	switch m.screen {
	case msg.ScreenEditor:
		return m.editor.View()
	default:
		return "Unknown screen"
	}
}
