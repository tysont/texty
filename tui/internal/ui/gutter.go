// ABOUTME: Renders the line number gutter for the editor.
// ABOUTME: Right-aligns line numbers with dimmed styling.

package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// GutterWidth returns the width of the gutter for a given line count.
func GutterWidth(totalLines int) int {
	w := len(fmt.Sprintf("%d", totalLines))
	if w < 3 {
		w = 3
	}
	return w + 1 // +1 for the separator space
}

// GutterLine renders a single gutter line number.
func GutterLine(lineNum, totalLines int) string {
	w := GutterWidth(totalLines) - 1 // minus separator
	num := fmt.Sprintf("%*d", w, lineNum)
	return lipgloss.NewStyle().
		Foreground(ColorTextDim).
		Render(num) +
		lipgloss.NewStyle().
			Foreground(ColorBorder).
			Render(" ")
}

// GutterEmpty renders an empty gutter line (for lines past the end of text).
func GutterEmpty(totalLines int) string {
	w := GutterWidth(totalLines) - 1
	return lipgloss.NewStyle().
		Foreground(ColorTextDim).
		Render(strings.Repeat(" ", w)) +
		lipgloss.NewStyle().
			Foreground(ColorBorder).
			Render(" ")
}
