// Package mailer provides a simple SMTP mailer for generated projects.
// Copy this package into your project — it does not import drp at runtime.
// Built on net/smtp only; no third-party mail library.
package mailer

import (
	"fmt"
	"net/smtp"
	"strings"
)

// Mailer sends transactional emails via SMTP.
type Mailer struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// New constructs a Mailer. Port is typically 587 (STARTTLS) or 465 (TLS).
func New(host string, port int, username, password, from string) *Mailer {
	return &Mailer{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
	}
}

// Send sends a plain-text email to the given recipient.
func (m *Mailer) Send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", m.Host, m.Port)
	auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)

	msg := buildMessage(m.From, to, subject, body)
	if err := smtp.SendMail(addr, auth, m.From, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("mailer: sending to %q: %w", to, err)
	}
	return nil
}

// SendHTML sends an HTML email to the given recipient.
func (m *Mailer) SendHTML(to, subject, htmlBody string) error {
	addr := fmt.Sprintf("%s:%d", m.Host, m.Port)
	auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)

	msg := buildHTMLMessage(m.From, to, subject, htmlBody)
	if err := smtp.SendMail(addr, auth, m.From, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("mailer: sending HTML to %q: %w", to, err)
	}
	return nil
}

func buildMessage(from, to, subject, body string) string {
	var sb strings.Builder
	sb.WriteString("From: " + from + "\r\n")
	sb.WriteString("To: " + to + "\r\n")
	sb.WriteString("Subject: " + subject + "\r\n")
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(body)
	return sb.String()
}

func buildHTMLMessage(from, to, subject, htmlBody string) string {
	var sb strings.Builder
	sb.WriteString("From: " + from + "\r\n")
	sb.WriteString("To: " + to + "\r\n")
	sb.WriteString("Subject: " + subject + "\r\n")
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(htmlBody)
	return sb.String()
}
