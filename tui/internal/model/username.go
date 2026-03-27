// ABOUTME: Username prompt screen shown on first launch.
// ABOUTME: Asks for a display name and saves it to the config file.

package model

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tysont/texty/tui/internal/config"
	"github.com/tysont/texty/tui/internal/ui"
)

// UsernameModel prompts for a username on first launch.
type UsernameModel struct {
	input  string
	width  int
	height int
}

func NewUsernameModel() UsernameModel {
	return UsernameModel{}
}

func (m UsernameModel) Init() tea.Cmd {
	return nil
}

func (m UsernameModel) Update(raw tea.Msg) (UsernameModel, tea.Cmd) {
	switch v := raw.(type) {
	case tea.WindowSizeMsg:
		m.width = v.Width
		m.height = v.Height
	case tea.KeyMsg:
		switch v.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			name := strings.TrimSpace(m.input)
			if name != "" {
				cfg := config.Load()
				cfg.Username = name
				_ = config.Save(cfg)
				return m, func() tea.Msg {
					return usernameSet{name: name}
				}
			}
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			if v.Type == tea.KeyRunes {
				m.input += string(v.Runes)
			}
		}
	}
	return m, nil
}

type usernameSet struct{ name string }

func (m UsernameModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(ui.ColorAccent).
		Bold(true)

	promptStyle := lipgloss.NewStyle().
		Foreground(ui.ColorText)

	inputStyle := lipgloss.NewStyle().
		Foreground(ui.ColorAccent)

	dimStyle := lipgloss.NewStyle().
		Foreground(ui.ColorTextDim)

	content := titleStyle.Render("texty") + "\n\n" +
		promptStyle.Render("What should we call you?") + "\n\n" +
		"  " + inputStyle.Render(m.input) + lipgloss.NewStyle().Reverse(true).Render(" ") + "\n\n" +
		dimStyle.Render("Press Enter to continue")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorBorder).
		Padding(2, 4).
		Render(content)

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		box)
}
