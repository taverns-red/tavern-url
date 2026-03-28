package email

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestNewSMTPSender(t *testing.T) {
	sender := NewSMTPSender("smtp.example.com", "587", "from@example.com", "secret")
	if sender == nil {
		t.Fatal("expected non-nil sender")
	}
	if sender.host != "smtp.example.com" {
		t.Errorf("expected host smtp.example.com, got %q", sender.host)
	}
	if sender.port != "587" {
		t.Errorf("expected port 587, got %q", sender.port)
	}
	if sender.from != "from@example.com" {
		t.Errorf("expected from from@example.com, got %q", sender.from)
	}
}

func TestNewNoopSender(t *testing.T) {
	sender := NewNoopSender()
	if sender == nil {
		t.Fatal("expected non-nil sender")
	}
}

func TestNoopSender_Send(t *testing.T) {
	sender := NewNoopSender()

	// Capture stdout to verify it logs.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := sender.Send("to@example.com", "Test Subject", "<p>Hello</p>")

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("NoopSender.Send should not return error, got: %v", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read pipe: %v", err)
	}
	output := buf.String()

	if output == "" {
		t.Error("expected NoopSender to log output")
	}
}

func TestSender_Interface(t *testing.T) {
	// Verify both types implement Sender interface.
	var _ Sender = (*SMTPSender)(nil)
	var _ Sender = (*NoopSender)(nil)
}
