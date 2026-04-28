package mcpclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_CreateSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/agents/designing-code/users/testuser/sessions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"id":      "session-123",
			"appName": "ClaudeAgent",
			"userId":  "testuser",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, http.DefaultClient)
	resp, err := c.CreateSession("designing-code", "testuser", "")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if resp.ID != "session-123" {
		t.Errorf("expected session-123, got %s", resp.ID)
	}
}

func TestClient_SendMessage_SSE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/agents/designing-code/run_sse" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		fmt.Fprint(w, "data: {\"id\":\"1\",\"author\":\"agent\",\"partial\":true,\"content\":{\"role\":\"model\",\"parts\":[{\"text\":\"partial\"}]}}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "data: {\"id\":\"2\",\"author\":\"agent\",\"turnComplete\":true,\"content\":{\"role\":\"model\",\"parts\":[{\"text\":\"final answer\"}]}}\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	c := NewClient(server.URL, http.DefaultClient)
	result, err := c.SendMessage("designing-code", "testuser", "session-123", "hello", "")
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}
	if result.Response != "final answer" {
		t.Errorf("expected 'final answer', got %q", result.Response)
	}
	if result.EventsCount != 2 {
		t.Errorf("expected 2 events, got %d", result.EventsCount)
	}
}

// Regression: a trailing harness/status event with non-empty text used to clobber
// the model's final answer because the bridge picked the last text-bearing event.
func TestClient_SendMessage_IgnoresTrailingStatusEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		fmt.Fprint(w, "data: {\"id\":\"1\",\"author\":\"agent\",\"partial\":true,\"content\":{\"role\":\"model\",\"parts\":[{\"text\":\"thinking...\"}]}}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "data: {\"id\":\"2\",\"author\":\"agent\",\"turnComplete\":true,\"content\":{\"role\":\"model\",\"parts\":[{\"text\":\"the actual review feedback\"}]}}\n\n")
		flusher.Flush()
		// Harness narration emitted after the turn completed — must not be picked.
		fmt.Fprint(w, "data: {\"id\":\"3\",\"author\":\"harness\",\"content\":{\"role\":\"model\",\"parts\":[{\"text\":\"LLM call started, iteration 5, 34 events\"}]}}\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	c := NewClient(server.URL, http.DefaultClient)
	result, err := c.SendMessage("reviewing-code", "testuser", "session-123", "review", "")
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}
	if result.Response != "the actual review feedback" {
		t.Errorf("expected final answer, got %q", result.Response)
	}
}

// When no event sets turnComplete, the bridge concatenates non-partial text events
// in order rather than dropping content.
func TestClient_SendMessage_NoTurnCompleteConcatenates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		fmt.Fprint(w, "data: {\"id\":\"1\",\"author\":\"agent\",\"content\":{\"role\":\"model\",\"parts\":[{\"text\":\"part one. \"}]}}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "data: {\"id\":\"2\",\"author\":\"agent\",\"content\":{\"role\":\"model\",\"parts\":[{\"text\":\"part two.\"}]}}\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	c := NewClient(server.URL, http.DefaultClient)
	result, err := c.SendMessage("reviewing-code", "testuser", "session-123", "go", "")
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}
	if result.Response != "part one. part two." {
		t.Errorf("expected concatenated text, got %q", result.Response)
	}
}

// Empty content on first attempt → retry once → succeed.
func TestClient_SendMessage_RetriesOnEmptyContent(t *testing.T) {
	var attempts int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		if attempts == 1 {
			// Stream contains only a process-narration event — no real content.
			fmt.Fprint(w, "data: {\"id\":\"1\",\"author\":\"agent\",\"partial\":true,\"content\":{\"role\":\"model\",\"parts\":[{\"text\":\"warming up\"}]}}\n\n")
			flusher.Flush()
			return
		}
		fmt.Fprint(w, "data: {\"id\":\"1\",\"author\":\"agent\",\"turnComplete\":true,\"content\":{\"role\":\"model\",\"parts\":[{\"text\":\"recovered answer\"}]}}\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	c := NewClient(server.URL, http.DefaultClient)
	result, err := c.SendMessage("reviewing-code", "testuser", "session-123", "go", "")
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts (1 retry), got %d", attempts)
	}
	if result.Response != "recovered answer" {
		t.Errorf("expected 'recovered answer', got %q", result.Response)
	}
}

// Empty content on both attempts → return error (no silent empty success).
func TestClient_SendMessage_FailsAfterOneRetry(t *testing.T) {
	var attempts int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		fmt.Fprint(w, "data: {\"id\":\"1\",\"author\":\"agent\",\"partial\":true,\"content\":{\"role\":\"model\",\"parts\":[{\"text\":\"still warming up\"}]}}\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	c := NewClient(server.URL, http.DefaultClient)
	_, err := c.SendMessage("reviewing-code", "testuser", "session-123", "go", "")
	if err == nil {
		t.Fatal("expected error after retry exhaustion, got nil")
	}
	if attempts != 2 {
		t.Errorf("expected exactly 2 attempts, got %d", attempts)
	}
}

func TestClient_ListSessions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]string{
			{"id": "s1", "appName": "ClaudeAgent", "userId": "testuser"},
			{"id": "s2", "appName": "ClaudeAgent", "userId": "testuser"},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, http.DefaultClient)
	sessions, err := c.ListSessions("designing-code", "testuser")
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(sessions))
	}
}

func TestClient_CreateSession_WithAgentName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["appName"] != "CustomAgent" {
			t.Errorf("expected CustomAgent, got %s", body["appName"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"id":      "session-123",
			"appName": "CustomAgent",
			"userId":  "testuser",
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, http.DefaultClient)
	resp, err := c.CreateSession("designing-code", "testuser", "CustomAgent")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if resp.AppName != "CustomAgent" {
		t.Errorf("expected CustomAgent, got %s", resp.AppName)
	}
}

func TestClient_SendMessage_WithAgentName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["appName"] != "CustomAgent" {
			t.Errorf("expected CustomAgent, got %s", body["appName"])
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"id\":\"1\",\"author\":\"agent\",\"turnComplete\":true,\"content\":{\"role\":\"model\",\"parts\":[{\"text\":\"done\"}]}}\n\n")
	}))
	defer server.Close()

	c := NewClient(server.URL, http.DefaultClient)
	_, err := c.SendMessage("designing-code", "testuser", "session-123", "hello", "CustomAgent")
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}
}

func TestClient_DeleteSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL, http.DefaultClient)
	err := c.DeleteSession("designing-code", "testuser", "session-123")
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}
}

