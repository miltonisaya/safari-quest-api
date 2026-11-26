package mailer

import (
	"strings"
	"testing"
)

func TestBuildMessageStripsHeaderInjection(t *testing.T) {
	msg := string(buildMessage("a@x.test", "b@x.test\r\nBcc: evil@x.test", "Hi\nBcc: evil@x.test", "line1\nline2"))
	head, body, _ := strings.Cut(msg, "\r\n\r\n")
	if strings.Contains(head, "\r\nBcc:") {
		t.Fatalf("header injection not stripped:\n%s", head)
	}
	if body != "line1\r\nline2" {
		t.Fatalf("body = %q", body)
	}
}

func TestBuildMessageEncodesNonASCIISubject(t *testing.T) {
	msg := string(buildMessage("a@x.test", "b@x.test", "Karibu Zürich", "x"))
	if !strings.Contains(msg, "Subject: =?UTF-8?q?") {
		t.Fatalf("subject not encoded: %s", msg)
	}
}
