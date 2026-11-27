package email

import (
	"strings"
	"testing"
	"time"
)

func TestContentRendersTextAndEscapedHTML(t *testing.T) {
	c := Content{
		Subject: "Your code", Heading: "Reset your password", Paragraphs: []string{"Use this code <now>."},
		Code: "123456", CodeNote: "Valid for 10 minutes.",
		Details: []Detail{{Label: "Reference", Value: "SQ-ABC123"}}, Quote: "Fully booked",
		Button: &Button{Label: "View booking", URL: "https://example.test/my-bookings"},
	}
	text := c.Text()
	for _, want := range []string{"Reset your password", "123456", "Reference:", "SQ-ABC123", "View booking: https://example.test/my-bookings"} {
		if !strings.Contains(text, want) {
			t.Errorf("text missing %q:\n%s", want, text)
		}
	}
	html, err := c.HTML()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "123456") || !strings.Contains(html, `href="https://example.test/my-bookings"`) {
		t.Fatalf("html missing code or button:\n%s", html)
	}
	if strings.Contains(html, "<now>") {
		t.Fatal("user text must be escaped in HTML")
	}
}

func TestRetryQueueNames(t *testing.T) {
	if got := retryQueue(5 * time.Minute); got != "emails.retry.300s" {
		t.Fatalf("retryQueue = %s", got)
	}
	if len(retryDelays) != MaxAttempts-1 {
		t.Fatalf("need a delay for each retry: %d delays, %d attempts", len(retryDelays), MaxAttempts)
	}
}
