// ABOUTME: Entry point for the Texty TUI collaborative text editor.
// ABOUTME: Parses flags and launches the Bubble Tea application.

package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"

	"github.com/google/uuid"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tysont/texty/tui/internal/model"
)

func main() {
	server := flag.String("server", "http://localhost:8787", "Backend server URL")
	username := flag.String("user", defaultUsername(), "Your display name")
	flag.Parse()

	userID := uuid.New().String()

	p := tea.NewProgram(
		model.NewAppModel(*server, userID, *username),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func defaultUsername() string {
	if u, err := user.Current(); err == nil && u.Username != "" {
		return u.Username
	}
	return "anonymous"
}
