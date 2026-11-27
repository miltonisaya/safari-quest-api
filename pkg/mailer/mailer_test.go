package mailer

import (
	"strings"
	"testing"
)

func TestBuildStripsHeaderInjection(t *testing.T) {
	msg := string(build("a@x.test", "b@x.test\r\nBcc: evil@x.test", Message{Subject: "Hi\nBcc: evil@x.test", Text: "line1\nline2"}))
	head, body, _ := strings.Cut(msg, "\r\n\r\n")
	if strings.Contains(head, "\r\nBcc:") {
		t.Fatalf("header injection not stripped:\n%s", head)
	}
	if !strings.Contains(body, "line1\r\nline2") {
		t.Fatalf("body = %q", body)
	}
}

func TestBuildEncodesNonASCIISubject(t *testing.T) {
	msg := string(build("a@x.test", "b@x.test", Message{Subject: "Karibu Zürich", Text: "x"}))
	if !strings.Contains(msg, "Subject: =?UTF-8?q?") {
		t.Fatalf("subject not encoded: %s", msg)
	}
}

func TestBuildHTMLWithTextAlternative(t *testing.T) {
	msg := string(build("a@x.test", "b@x.test", Message{Subject: "S", Text: "Plain", HTML: "<p>Rich</p>"}))
	if !strings.Contains(msg, "multipart/alternative") || !strings.Contains(msg, "text/plain") || !strings.Contains(msg, "text/html") {
		t.Fatalf("missing parts:\n%s", msg)
	}
	if strings.Index(msg, "text/plain") > strings.Index(msg, "text/html") {
		t.Fatal("plain text must come first so clients prefer HTML")
	}
}
