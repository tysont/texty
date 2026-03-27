// ABOUTME: SSE (Server-Sent Events) stream reader for real-time updates.
// ABOUTME: Connects to the backend subscribe endpoint and emits Bubble Tea messages.

package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tysont/texty/tui/internal/msg"
)

// ListenSSE returns a Bubble Tea command that blocks reading SSE events.
// It sends SSEUpdate messages for each event and SSEError on failure.
// Cancel the context to stop listening. url should be the full subscribe URL.
func ListenSSE(ctx context.Context, url string) tea.Cmd {
	return func() tea.Msg {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return msg.SSEError{Err: fmt.Errorf("create SSE request: %w", err)}
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return nil // context cancelled, not an error
			}
			return msg.SSEError{Err: fmt.Errorf("SSE connect: %w", err)}
		}
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		var dataLine string

		for scanner.Scan() {
			if ctx.Err() != nil {
				return nil
			}

			line := scanner.Text()

			if strings.HasPrefix(line, "data: ") {
				dataLine = strings.TrimPrefix(line, "data: ")
			} else if line == "" && dataLine != "" {
				var update struct {
					Text           string   `json:"text"`
					LockHolder     string   `json:"lockHolder"`
					LockHolderName string   `json:"lockHolderName"`
					Users          []string `json:"users"`
				}
				if err := json.Unmarshal([]byte(dataLine), &update); err == nil {
					return msg.SSEUpdate{
						Text:           update.Text,
						LockHolder:     update.LockHolder,
						LockHolderName: update.LockHolderName,
						Users:          update.Users,
					}
				}
				dataLine = ""
			}
		}

		if err := scanner.Err(); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return msg.SSEError{Err: fmt.Errorf("SSE read: %w", err)}
		}

		return msg.SSEError{Err: fmt.Errorf("SSE stream ended")}
	}
}
