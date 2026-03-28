package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeliver_Success(t *testing.T) {
	// Create a test server that accepts webhook POSTs.
	received := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = true
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if sig := r.Header.Get("X-Tavern-Signature"); sig == "" {
			t.Error("expected X-Tavern-Signature header")
		}
		if event := r.Header.Get("X-Tavern-Event"); event != "link.created" {
			t.Errorf("expected event link.created, got %q", event)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := NewWebhookService()
	err := svc.Deliver(server.URL, "test-secret", "link.created", map[string]string{"slug": "abc"})
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if !received {
		t.Error("webhook was not delivered")
	}
}

func TestDeliver_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	svc := NewWebhookService()
	err := svc.Deliver(server.URL, "test-secret", "link.deleted", nil)
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestDeliver_InvalidURL(t *testing.T) {
	svc := NewWebhookService()
	err := svc.Deliver("http://localhost:1/nonexistent", "secret", "test", nil)
	if err == nil {
		t.Error("expected error for unreachable server")
	}
}

func TestWebhookPayload_Structure(t *testing.T) {
	// Verify payload can be marshaled correctly.
	payload := WebhookPayload{
		Event:     "link.created",
		Timestamp: "2026-03-28T10:00:00Z",
		Data:      map[string]string{"slug": "abc"},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if decoded["event"] != "link.created" {
		t.Errorf("expected event link.created, got %v", decoded["event"])
	}
}
