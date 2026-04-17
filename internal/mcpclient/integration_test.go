package mcpclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIntegration_CreateSession_SendMessage(t *testing.T) {
	// Mock gateway that implements the same HTTP API surface.
	gw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		switch {
		// POST /v1/agents/{role}/users/{userID}/sessions → create session
		case r.Method == "POST" && strings.HasSuffix(path, "/users/testuser/sessions"):
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"id":      "sess-abc",
				"appName": "ClaudeAgent",
				"userId":  "testuser",
			})

		// POST /v1/agents/{role}/run_sse → send message (SSE)
		case r.Method == "POST" && strings.HasSuffix(path, "/run_sse"):
			w.Header().Set("Content-Type", "text/event-stream")
			flusher := w.(http.Flusher)
			fmt.Fprint(w, "data: {\"id\":\"e1\",\"author\":\"ClaudeAgent\",\"partial\":true,\"content\":{\"role\":\"model\",\"parts\":[{\"text\":\"thinking...\"}]}}\n\n")
			flusher.Flush()
			fmt.Fprint(w, "data: {\"id\":\"e2\",\"author\":\"ClaudeAgent\",\"turnComplete\":true,\"content\":{\"role\":\"model\",\"parts\":[{\"text\":\"Here is the answer.\"}]}}\n\n")
			flusher.Flush()

		// GET /v1/agents/{role}/users/{userID}/sessions → list sessions
		case r.Method == "GET" && strings.HasSuffix(path, "/users/testuser/sessions"):
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]map[string]any{
				{"id": "sess-abc", "appName": "ClaudeAgent", "userId": "testuser"},
			})

		// DELETE /v1/agents/{role}/users/{userID}/sessions/{sessionID}
		case r.Method == "DELETE" && strings.HasSuffix(path, "/sessions/sess-abc"):
			w.WriteHeader(http.StatusOK)

		default:
			http.Error(w, "not found: "+path, 404)
		}
	}))
	defer gw.Close()

	// MCP client talking to mock gateway.
	client := NewClient(gw.URL, http.DefaultClient)

	// 1. Create session.
	sess, err := client.CreateSession("coding-design", "testuser", "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if sess.ID != "sess-abc" {
		t.Errorf("expected sess-abc, got %s", sess.ID)
	}

	// 2. Send message (streaming).
	result, err := client.SendMessage("coding-design", "testuser", sess.ID, "What is 2+2?", "")
	if err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	if result.Response != "Here is the answer." {
		t.Errorf("expected 'Here is the answer.', got %q", result.Response)
	}
	if result.EventsCount != 2 {
		t.Errorf("expected 2 events, got %d", result.EventsCount)
	}

	// 3. List sessions.
	sessions, err := client.ListSessions("coding-design", "testuser")
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Errorf("expected 1 session, got %d", len(sessions))
	}

	// 4. Delete session.
	err = client.DeleteSession("coding-design", "testuser", "sess-abc")
	if err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}
}
