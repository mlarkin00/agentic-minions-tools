package mcpclient

import (
	"strings"
	"testing"
)

func TestReadSSEEvents(t *testing.T) {
	stream := strings.NewReader(
		"data: {\"id\":\"1\",\"author\":\"agent\",\"partial\":true}\n\n" +
			"data: {\"id\":\"2\",\"author\":\"agent\",\"turnComplete\":true}\n\n",
	)

	events, err := ReadSSEEvents(stream)
	if err != nil {
		t.Fatalf("ReadSSEEvents failed: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].ID != "1" {
		t.Errorf("event 0 id: expected 1, got %s", events[0].ID)
	}
	if !events[1].TurnComplete {
		t.Error("event 1 should have turnComplete=true")
	}
}

func TestReadSSEEvents_Empty(t *testing.T) {
	stream := strings.NewReader("")
	events, err := ReadSSEEvents(stream)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestSSEEvent_Text(t *testing.T) {
	stream := strings.NewReader(
		"data: {\"id\":\"1\",\"content\":{\"role\":\"model\",\"parts\":[{\"text\":\"hello \"},{\"text\":\"world\"}]}}\n\n",
	)
	events, err := ReadSSEEvents(stream)
	if err != nil {
		t.Fatalf("ReadSSEEvents failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Text() != "hello world" {
		t.Errorf("expected 'hello world', got %q", events[0].Text())
	}
}

func TestSSEEvent_Text_NoContent(t *testing.T) {
	e := SSEEvent{ID: "1"}
	if e.Text() != "" {
		t.Errorf("expected empty, got %q", e.Text())
	}
}
