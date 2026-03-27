// ABOUTME: Entry point for the Texty TUI collaborative text editor.
// ABOUTME: Initializes Bubble Tea and launches the application.

package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tysont/texty/tui/internal/model"
)

func main() {
	p := tea.NewProgram(
		model.NewAppModel(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
