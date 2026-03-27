// ABOUTME: Renders the status bar at the bottom of the editor.
// ABOUTME: Shows cursor position, keybind hints, editing mode, and connection state.

package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// StatusBar renders the bottom bar of the editor.
func StatusBar(width, row, col int, editing, connected bool) string {
	style := lipgloss.NewStyle().
		Background(ColorSurface).
		Foreground(ColorTextDim).
		Padding(0, 1).
		Width(width)

	pos := fmt.Sprintf("Ln %d, Col %d", row+1, col+1)

	hints := "?: help" + Dot + "^S: save" + Dot + "^Q: back"

	modeStyle := lipgloss.NewStyle().Background(ColorSurface)
	var mode string
	if !connected {
		mode = modeStyle.Foreground(ColorError).Render("disconnected")
	} else if editing {
		mode = modeStyle.Foreground(ColorSuccess).Render("editing")
	} else {
		mode = modeStyle.Foreground(ColorTextDim).Render("viewing")
	}

	left := pos + Dot + hints
	right := mode

	gap := width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}

	spacer := lipgloss.NewStyle().
		Background(ColorSurface).
		Width(gap).
		Render("")

	return style.Render(left + spacer + right)
}

// DocListStatusBar renders the bottom bar for the document list screen.
func DocListStatusBar(width int) string {
	style := lipgloss.NewStyle().
		Background(ColorSurface).
		Foreground(ColorTextDim).
		Padding(0, 1).
		Width(width)

	return style.Render("enter: open" + Dot + "n: new" + Dot + "d: delete" + Dot + "r: refresh" + Dot + "q: quit")
}
