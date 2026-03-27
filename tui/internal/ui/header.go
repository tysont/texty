// ABOUTME: Renders the header bar at the top of the editor and document list.
// ABOUTME: Shows app name, document name, connected users, and lock status.

package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// HeaderBar renders the top bar of the application.
func HeaderBar(width int, parts ...string) string {
	style := lipgloss.NewStyle().
		Background(ColorSurface).
		Foreground(ColorText).
		Padding(0, 1).
		Width(width)

	content := ""
	for i, p := range parts {
		if i > 0 {
			content += lipgloss.NewStyle().
				Foreground(ColorTextDim).
				Background(ColorSurface).
				Render(Dot)
		}
		content += p
	}

	return style.Render(content)
}

// HeaderAccent renders a piece of header text with the accent color.
func HeaderAccent(s string) string {
	return lipgloss.NewStyle().
		Foreground(ColorAccent).
		Background(ColorSurface).
		Bold(true).
		Render(s)
}

// HeaderConnected renders the connected user count.
func HeaderConnected(n int) string {
	label := "connected"
	if n == 1 {
		label = "connected"
	}
	return lipgloss.NewStyle().
		Foreground(ColorTextDim).
		Background(ColorSurface).
		Render(fmt.Sprintf("%d %s", n, label))
}

// HeaderLock renders the lock holder display.
func HeaderLock(username string) string {
	if username == "" {
		return lipgloss.NewStyle().
			Foreground(ColorTextDim).
			Background(ColorSurface).
			Render("unlocked")
	}
	return lipgloss.NewStyle().
		Foreground(ColorWarning).
		Background(ColorSurface).
		Render("lock: @" + username)
}
