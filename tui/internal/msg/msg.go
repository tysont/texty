// ABOUTME: Defines all Bubble Tea message types used across models.
// ABOUTME: Centralizes message definitions for clean inter-model communication.

package msg

import tea "github.com/charmbracelet/bubbletea"

// ScreenType identifies which screen is active.
type ScreenType int

const (
	ScreenUsername ScreenType = iota
	ScreenDocList
	ScreenEditor
)

// SwitchScreen transitions to a different screen.
type SwitchScreen struct {
	Screen ScreenType
	DocID  string // populated when switching to Editor
}

// SSEUpdate arrives from the server via SSE.
type SSEUpdate struct {
	Text       string
	LockHolder string
	Users      []string
}

// SSEError indicates the SSE connection failed.
type SSEError struct {
	Err error
}

// LockAcquired is the response from a lock acquire attempt.
type LockAcquired struct {
	Success bool
}

// LockReleased confirms lock release.
type LockReleased struct{}

// TextSaved confirms text was saved.
type TextSaved struct{}

// DocsListed returns the list of documents.
type DocsListed struct {
	Docs []DocSummary
}

// DocSummary describes a document in the list.
type DocSummary struct {
	ID             string
	Name           string
	ConnectedUsers int
	LineCount      int
}

// DocCreated confirms a new document was created.
type DocCreated struct {
	ID string
}

// DocDeleted confirms a document was deleted.
type DocDeleted struct {
	ID string
}

// SaveTick fires when the save debounce timer expires.
type SaveTick struct {
	Gen int
}

// IdleTimeout fires when the idle lock release timer expires.
type IdleTimeout struct {
	Gen int
}

// ForceQuit exits the application.
type ForceQuit struct{}

// Quit returns a quit command.
func Quit() tea.Msg {
	return ForceQuit{}
}
