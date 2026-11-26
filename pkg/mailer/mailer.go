// Package mailer sends plain-text email over SMTP.
package mailer

import (
	"fmt"
	"log"
	"mime"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"safari-quest-api/config"
)

// Send emails one recipient. Without SMTP_HOST it logs the message instead,
// so flows like password reset can be tested locally.
// shortcut: STARTTLS on port 587 (or plain on 25) only, no implicit TLS on 465; add it if a provider needs it.
func Send(to, subject, body string) error {
	cfg := config.App
	if cfg.SMTPHost == "" {
		log.Printf("[mailer] SMTP_HOST not set; email not sent.\nTo: %s\nSubject: %s\n\n%s\n", to, subject, body)
		return nil
	}
	from, err := mail.ParseAddress(cfg.MailFrom)
	if err != nil {
		return fmt.Errorf("invalid MAIL_FROM: %w", err)
	}
	msg := buildMessage(from.String(), to, subject, body)
	var auth smtp.Auth
	if cfg.SMTPUsername != "" {
		auth = smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPHost)
	}
	return smtp.SendMail(cfg.SMTPHost+":"+cfg.SMTPPort, auth, from.Address, []string{to}, msg)
}

func buildMessage(from, to, subject, body string) []byte {
	// Header values must not contain line breaks (header injection).
	clean := func(s string) string { return strings.NewReplacer("\r", "", "\n", "").Replace(s) }
	headers := []string{
		"From: " + clean(from),
		"To: " + clean(to),
		"Subject: " + mime.QEncoding.Encode("UTF-8", clean(subject)), // encodes only if non-ASCII
		"Date: " + time.Now().Format(time.RFC1123Z),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
	}
	return []byte(strings.Join(headers, "\r\n") + "\r\n\r\n" + strings.ReplaceAll(body, "\n", "\r\n"))
}
