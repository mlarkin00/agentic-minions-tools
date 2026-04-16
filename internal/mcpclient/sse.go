package mcpclient

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// SSEEvent represents a parsed event from the ADK SSE stream.
type SSEEvent struct {
	ID           string `json:"id"`
	Author       string `json:"author"`
	Partial      bool   `json:"partial,omitempty"`
	TurnComplete bool   `json:"turnComplete,omitempty"`
	ErrorCode    string `json:"errorCode,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	Content      *struct {
		Role  string `json:"role"`
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"content,omitempty"`
}

// Text returns the concatenated text from all parts, or empty string.
func (e SSEEvent) Text() string {
	if e.Content == nil {
		return ""
	}
	var sb strings.Builder
	for _, p := range e.Content.Parts {
		sb.WriteString(p.Text)
	}
	return sb.String()
}

// ReadSSEEvents reads all SSE events from a stream until EOF.
func ReadSSEEvents(r io.Reader) ([]SSEEvent, error) {
	var events []SSEEvent
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		var event SSEEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return events, fmt.Errorf("parse SSE event: %w", err)
		}
		events = append(events, event)
	}
	return events, scanner.Err()
}
