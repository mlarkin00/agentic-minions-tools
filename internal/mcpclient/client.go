package mcpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client communicates with the gateway API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a Client for the given gateway URL with the provided HTTP client.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	return &Client{baseURL: baseURL, httpClient: httpClient}
}

// SessionResponse is the response from session create/get operations.
type SessionResponse struct {
	ID      string `json:"id"`
	AppName string `json:"appName"`
	UserID  string `json:"userId"`
}

// SendMessageResult is the result from sending a message.
type SendMessageResult struct {
	Response    string `json:"response"`
	EventsCount int    `json:"events_count"`
}

// CreateSession creates a new session for the given role and user.
func (c *Client) CreateSession(role, userID, agentName string) (*SessionResponse, error) {
	url := fmt.Sprintf("%s/v1/agents/%s/users/%s/sessions", c.baseURL, role, userID)

	if agentName == "" {
		agentName = "ClaudeAgent"
	}
	bodyBytes, _ := json.Marshal(map[string]string{"appName": agentName})

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create session request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create session failed (%d): %s", resp.StatusCode, body)
	}

	var session SessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("decode session response: %w", err)
	}
	return &session, nil
}

// SendMessage sends a message to an agent via SSE and returns the final response.
//
// If the agent returns no content (empty response after stripping process/status events),
// the request is retried once on the same session. Two total attempts; failure thereafter
// is surfaced as an error so callers can't accidentally treat an empty response as success.
func (c *Client) SendMessage(role, userID, sessionID, message, agentName string) (*SendMessageResult, error) {
	const maxAttempts = 2 // initial attempt + 1 retry
	var lastResult *SendMessageResult
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err := c.sendMessageOnce(role, userID, sessionID, message, agentName)
		if err != nil {
			return nil, err
		}
		if result.Response != "" {
			return result, nil
		}
		lastResult = result
	}
	return nil, fmt.Errorf("agent returned no content after %d attempts (events_count=%d)", maxAttempts, lastResult.EventsCount)
}

func (c *Client) sendMessageOnce(role, userID, sessionID, message, agentName string) (*SendMessageResult, error) {
	url := fmt.Sprintf("%s/v1/agents/%s/run_sse", c.baseURL, role)

	if agentName == "" {
		agentName = "ClaudeAgent"
	}

	body := map[string]any{
		"appName":   agentName,
		"userId":    userID,
		"sessionId": sessionID,
		"newMessage": map[string]any{
			"role":  "user",
			"parts": []map[string]string{{"text": message}},
		},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("send message request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("send message failed (%d): %s", resp.StatusCode, errBody)
	}

	events, err := ReadSSEEvents(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read SSE events: %w", err)
	}

	return &SendMessageResult{
		Response:    selectFinalText(events),
		EventsCount: len(events),
	}, nil
}

// ListSessions lists sessions for a role and user.
func (c *Client) ListSessions(role, userID string) ([]SessionResponse, error) {
	url := fmt.Sprintf("%s/v1/agents/%s/users/%s/sessions", c.baseURL, role, userID)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("list sessions request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list sessions failed (%d): %s", resp.StatusCode, body)
	}

	var sessions []SessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		return nil, fmt.Errorf("decode sessions: %w", err)
	}
	return sessions, nil
}

// DeleteSession deletes a session.
func (c *Client) DeleteSession(role, userID, sessionID string) error {
	url := fmt.Sprintf("%s/v1/agents/%s/users/%s/sessions/%s", c.baseURL, role, userID, sessionID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("create delete request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete session request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete session failed (%d): %s", resp.StatusCode, body)
	}
	return nil
}
