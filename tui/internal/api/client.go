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

// GetText fetches the current text and lock state for a document.
func (c *Client) GetText(docID string) (*TextState, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/docs/" + docID + "/text")
	if err != nil {
		return nil, fmt.Errorf("GET text: %w", err)
	}
	defer resp.Body.Close()

	var state TextState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, fmt.Errorf("decode text: %w", err)
	}
	return &state, nil
}

// PostText updates the text on the server. Only succeeds if userId holds the lock.
func (c *Client) PostText(docID, userId, text string) error {
	body, _ := json.Marshal(map[string]string{
		"userId": userId,
		"text":   text,
	})
	resp, err := c.HTTPClient.Post(c.BaseURL+"/docs/"+docID+"/text", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("POST text: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("POST text: status %d", resp.StatusCode)
	}
	return nil
}

// AcquireLock attempts to acquire the editing lock for a document.
func (c *Client) AcquireLock(docID, userId string) (bool, error) {
	body, _ := json.Marshal(map[string]string{"userId": userId})
	resp, err := c.HTTPClient.Post(c.BaseURL+"/docs/"+docID+"/lock/acquire", "application/json", bytes.NewReader(body))
	if err != nil {
		return false, fmt.Errorf("POST lock/acquire: %w", err)
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

// ReleaseLock releases the editing lock for a document.
func (c *Client) ReleaseLock(docID, userId string) error {
	body, _ := json.Marshal(map[string]string{"userId": userId})
	resp, err := c.HTTPClient.Post(c.BaseURL+"/docs/"+docID+"/lock/release", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("POST lock/release: %w", err)
	}
	defer resp.Body.Close()
	return nil
}

// SubscribeURL returns the SSE subscribe URL for a document with user info.
func (c *Client) SubscribeURL(docID, userID, username string) string {
	return fmt.Sprintf("%s/docs/%s/subscribe?userId=%s&username=%s", c.BaseURL, docID, userID, username)
}

// DocSummary is a document in the list response.
type DocSummary struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	ConnectedUsers int    `json:"connectedUsers"`
	LineCount      int    `json:"lineCount"`
}

// ListDocs fetches the list of all documents.
func (c *Client) ListDocs() ([]DocSummary, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/docs")
	if err != nil {
		return nil, fmt.Errorf("GET /docs: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Docs []DocSummary `json:"docs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode GET /docs: %w", err)
	}
	return result.Docs, nil
}

// CreateDoc creates a new document and returns its ID.
func (c *Client) CreateDoc(name string) (string, error) {
	body, _ := json.Marshal(map[string]string{"name": name})
	resp, err := c.HTTPClient.Post(c.BaseURL+"/docs", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("POST /docs: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		ID    string `json:"id"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode POST /docs: %w", err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("create doc: %s", result.Error)
	}
	return result.ID, nil
}

// DeleteDoc deletes a document by ID.
func (c *Client) DeleteDoc(docID string) error {
	req, err := http.NewRequest("DELETE", c.BaseURL+"/docs/"+docID, nil)
	if err != nil {
		return fmt.Errorf("DELETE /docs/%s: %w", docID, err)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("DELETE /docs/%s: %w", docID, err)
	}
	defer resp.Body.Close()
	return nil
}
