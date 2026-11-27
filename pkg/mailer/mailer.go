// Package mailer sends email over SMTP: plain text, or HTML with a plain-text
// alternative.
package mailer

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"safari-quest-api/config"
)

// ErrNotConfigured: SMTP_HOST is empty, so nothing can be sent. The email
// worker treats it like any failed send and retries later.
var ErrNotConfigured = errors.New("SMTP_HOST is not set")

// Message is one email. HTML is optional; Text is always sent.
type Message struct {
	To      string
	Subject string
	Text    string
	HTML    string
}

// Send emails one recipient a plain-text message.
func Send(to, subject, body string) error {
	return SendMessage(Message{To: to, Subject: subject, Text: body})
}

// SendMessage delivers m over SMTP.
// shortcut: STARTTLS on port 587 (or plain on 25/1025) only, no implicit TLS on 465; add it if a provider needs it.
func SendMessage(m Message) error {
	cfg := config.App
	if cfg.SMTPHost == "" {
		return ErrNotConfigured
	}
	from, err := mail.ParseAddress(cfg.MailFrom)
	if err != nil {
		return fmt.Errorf("invalid MAIL_FROM: %w", err)
	}
	to, err := mail.ParseAddress(m.To)
	if err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}
	var auth smtp.Auth
	if cfg.SMTPUsername != "" {
		auth = smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPHost)
	}
	return smtp.SendMail(cfg.SMTPHost+":"+cfg.SMTPPort, auth, from.Address, []string{to.Address}, build(from.String(), to.Address, m))
}

// clean strips line breaks from header values (header injection).
func clean(s string) string { return strings.NewReplacer("\r", "", "\n", "").Replace(s) }

func qp(s string) string {
	var b strings.Builder
	w := quotedprintable.NewWriter(&b)
	_, _ = w.Write([]byte(strings.ReplaceAll(s, "\n", "\r\n")))
	_ = w.Close()
	return b.String()
}

func build(from, to string, m Message) []byte {
	headers := []string{
		"From: " + clean(from),
		"To: " + clean(to),
		"Subject: " + mime.QEncoding.Encode("UTF-8", clean(m.Subject)), // encodes only if non-ASCII
		"Date: " + time.Now().Format(time.RFC1123Z),
		"MIME-Version: 1.0",
	}
	if m.HTML == "" {
		headers = append(headers, "Content-Type: text/plain; charset=UTF-8", "Content-Transfer-Encoding: quoted-printable")
		return []byte(strings.Join(headers, "\r\n") + "\r\n\r\n" + qp(m.Text))
	}
	random := make([]byte, 12)
	_, _ = rand.Read(random)
	boundary := "sq-" + hex.EncodeToString(random)
	headers = append(headers, `Content-Type: multipart/alternative; boundary="`+boundary+`"`)
	var b strings.Builder
	b.WriteString(strings.Join(headers, "\r\n") + "\r\n\r\n")
	for _, part := range []struct{ kind, body string }{{"text/plain", m.Text}, {"text/html", m.HTML}} {
		b.WriteString("--" + boundary + "\r\n")
		b.WriteString("Content-Type: " + part.kind + "; charset=UTF-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n")
		b.WriteString(qp(part.body) + "\r\n")
	}
	b.WriteString("--" + boundary + "--\r\n")
	return []byte(b.String())
}
