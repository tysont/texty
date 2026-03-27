// ABOUTME: HTTP client for the Texty backend REST API.
// ABOUTME: Wraps all API calls (text, lock) with typed request/response structs.

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Client communicates with the Texty backend.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new API client.
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

// TextState is the response from GET /text.
type TextState struct {
	Text       string `json:"text"`
	LockHolder string `json:"lockHolder"`
}

// GetText fetches the current text and lock state.
func (c *Client) GetText() (*TextState, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/text")
	if err != nil {
		return nil, fmt.Errorf("GET /text: %w", err)
	}
	defer resp.Body.Close()

	var state TextState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, fmt.Errorf("decode GET /text: %w", err)
	}
	return &state, nil
}

// PostText updates the text on the server. Only succeeds if userId holds the lock.
func (c *Client) PostText(userId, text string) error {
	body, _ := json.Marshal(map[string]string{
		"userId": userId,
		"text":   text,
	})
	resp, err := c.HTTPClient.Post(c.BaseURL+"/text", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("POST /text: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("POST /text: status %d", resp.StatusCode)
	}
	return nil
}

// AcquireLock attempts to acquire the editing lock.
func (c *Client) AcquireLock(userId string) (bool, error) {
	body, _ := json.Marshal(map[string]string{"userId": userId})
	resp, err := c.HTTPClient.Post(c.BaseURL+"/lock/acquire", "application/json", bytes.NewReader(body))
	if err != nil {
		return false, fmt.Errorf("POST /lock/acquire: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("decode lock/acquire: %w", err)
	}
	return result.Success, nil
}

// ReleaseLock releases the editing lock.
func (c *Client) ReleaseLock(userId string) error {
	body, _ := json.Marshal(map[string]string{"userId": userId})
	resp, err := c.HTTPClient.Post(c.BaseURL+"/lock/release", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("POST /lock/release: %w", err)
	}
	defer resp.Body.Close()
	return nil
}
