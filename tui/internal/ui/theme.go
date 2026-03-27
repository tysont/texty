// ABOUTME: Defines the color palette and Lip Gloss styles for the TUI.
// ABOUTME: All visual styling is centralized here for consistency.

package ui

import "github.com/charmbracelet/lipgloss"

// Colors — dark theme with purple accent.
var (
	ColorBg        = lipgloss.Color("#121212")
	ColorSurface   = lipgloss.Color("#1e1e1e")
	ColorText      = lipgloss.Color("#e0e0e0")
	ColorTextDim   = lipgloss.Color("#666666")
	ColorAccent    = lipgloss.Color("#7c3aed")
	ColorAccentDim = lipgloss.Color("#4c1d95")
	ColorSuccess   = lipgloss.Color("#22c55e")
	ColorWarning   = lipgloss.Color("#eab308")
	ColorError     = lipgloss.Color("#ef4444")
	ColorBorder    = lipgloss.Color("#333333")
)

// Dot is the separator used in header/status bars.
const Dot = " · "
