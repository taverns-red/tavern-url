package email

import (
	"fmt"
	"net/smtp"
)

// Sender defines the interface for sending emails.
type Sender interface {
	Send(to, subject, body string) error
}

// SMTPSender sends emails via SMTP.
type SMTPSender struct {
	host     string
	port     string
	from     string
	password string
}

// NewSMTPSender creates a new SMTPSender.
func NewSMTPSender(host, port, from, password string) *SMTPSender {
	return &SMTPSender{host: host, port: port, from: from, password: password}
}

// Send sends an email via SMTP.
func (s *SMTPSender) Send(to, subject, body string) error {
	auth := smtp.PlainAuth("", s.from, s.password, s.host)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s.from, to, subject, body)
	return smtp.SendMail(s.host+":"+s.port, auth, s.from, []string{to}, []byte(msg))
}

// NoopSender is a no-op email sender for development.
type NoopSender struct{}

// NewNoopSender creates a no-op email sender.
func NewNoopSender() *NoopSender { return &NoopSender{} }

// Send logs the email but doesn't actually send it.
func (n *NoopSender) Send(to, subject, body string) error {
	fmt.Printf("[EMAIL] To: %s, Subject: %s\n", to, subject)
	return nil
}
