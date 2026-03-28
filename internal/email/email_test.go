package email

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
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

func TestSMTPSender_Send_ConnectionError(t *testing.T) {
	// Use a non-routable address to exercise all Send() code paths
	// (auth creation, message formatting) without an SMTP server.
	sender := NewSMTPSender("127.0.0.1", "19999", "test@test.com", "pass")
	err := sender.Send("to@test.com", "Subject", "<p>Body</p>")
	if err == nil {
		t.Error("expected connection error when no SMTP server running")
	}
}

func TestSMTPSender_Send_WithMockServer(t *testing.T) {
	// Start a minimal TCP listener acting as a mock SMTP server.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()
	addr := ln.Addr().(*net.TCPAddr) //nolint:errcheck
	port := addr.Port

	// Run a tiny goroutine that does the bare SMTP handshake.
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// SMTP greeting.
		fmt.Fprintf(conn, "220 mock SMTP\r\n")

		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			line := scanner.Text()
			switch {
			case len(line) >= 4 && line[:4] == "EHLO":
				fmt.Fprintf(conn, "250-mock Hello\r\n250 AUTH PLAIN LOGIN\r\n")
			case len(line) >= 4 && line[:4] == "HELO":
				fmt.Fprintf(conn, "250 Hello\r\n")
			case len(line) >= 4 && line[:4] == "AUTH":
				fmt.Fprintf(conn, "235 Authentication successful\r\n")
			case len(line) >= 4 && line[:4] == "MAIL":
				fmt.Fprintf(conn, "250 OK\r\n")
			case len(line) >= 4 && line[:4] == "RCPT":
				fmt.Fprintf(conn, "250 OK\r\n")
			case len(line) >= 4 && line[:4] == "DATA":
				fmt.Fprintf(conn, "354 Start mail input\r\n")
			case line == ".":
				fmt.Fprintf(conn, "250 OK\r\n")
			case len(line) >= 4 && line[:4] == "QUIT":
				fmt.Fprintf(conn, "221 Bye\r\n")
				return
			}
		}
	}()

	sender := NewSMTPSender("127.0.0.1", fmt.Sprintf("%d", port), "from@test.com", "secret")
	err = sender.Send("to@test.com", "Test Subject", "<p>Hello</p>")
	if err != nil {
		t.Errorf("expected successful send with mock server, got: %v", err)
	}
}

