// ABOUTME: Renders the help overlay showing keyboard shortcuts.
// ABOUTME: Displayed as a centered box over the editor content.

package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var helpEntries = []struct {
	key  string
	desc string
}{
	{"arrows", "move cursor"},
	{"Home/End", "start/end of line"},
	{"^A/^E", "start/end of line"},
	{"Enter", "new line"},
	{"Backspace", "delete before cursor"},
	{"Delete", "delete at cursor"},
	{"^S", "force save"},
	{"^Q", "back to doc list"},
	{"?", "toggle this help"},
}

// HelpOverlay renders the help box centered within the given dimensions.
func HelpOverlay(width, height int) string {
	boxWidth := 36

	titleStyle := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true)

	keyStyle := lipgloss.NewStyle().
		Foreground(ColorText).
		Width(12)

	descStyle := lipgloss.NewStyle().
		Foreground(ColorTextDim)

	content := titleStyle.Render("  Keyboard Shortcuts") + "\n\n"
	for _, e := range helpEntries {
		content += "  " + keyStyle.Render(e.key) + descStyle.Render(e.desc) + "\n"
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccent).
		Background(ColorSurface).
		Padding(1, 1).
		Width(boxWidth).
		Render(content)

	return lipgloss.Place(width, height,
		lipgloss.Center, lipgloss.Center,
		box)
}
