package mcpclient

import (
	"testing"
)

func TestNewAuthClient_DevMode(t *testing.T) {
	client, err := NewAuthClient("")
	if err != nil {
		t.Fatalf("NewAuthClient failed: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}
