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

// selectFinalText extracts the agent's final response text from a stream of SSE events.
//
// The ADK SSE stream interleaves model output with harness/tool narration. Picking the
// last event with non-empty text is wrong: a trailing harness status event (e.g. progress
// narration emitted after the model turn ends) will outrank the actual answer. We instead:
//  1. Prefer the last event explicitly marked turnComplete (and not a partial chunk).
//  2. Fall back to concatenating non-partial text events in order, for streams that don't
//     set turnComplete on a content-bearing event.
//
// A stream containing only partial chunks (or only events with empty text) returns "" —
// the caller is expected to treat that as an incomplete response and retry.
func selectFinalText(events []SSEEvent) string {
	for i := len(events) - 1; i >= 0; i-- {
		e := events[i]
		if e.TurnComplete && !e.Partial {
			if t := e.Text(); t != "" {
				return t
			}
		}
	}
	var sb strings.Builder
	for _, e := range events {
		if e.Partial {
			continue
		}
		if t := e.Text(); t != "" {
			sb.WriteString(t)
		}
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
