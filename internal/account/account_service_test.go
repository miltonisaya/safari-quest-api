package account

import (
	"strings"
	"testing"
)

func TestTokensAreRandomAndHashed(t *testing.T) {
	a, err := newToken()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := newToken()
	if a == b || len(a) < 40 {
		t.Fatalf("tokens should be long and unique: %q %q", a, b)
	}
	if h := hashToken(a); len(h) != 64 || h == a || h != hashToken(a) {
		t.Fatalf("hash should be a stable 64-char hex digest, got %q", h)
	}
}

func TestResetEmailContainsLinkAndExpiry(t *testing.T) {
	body := resetEmail("Asha", "http://localhost:4200/reset-password?token=abc")
	for _, want := range []string{"Hello Asha", "reset-password?token=abc", "60 minutes"} {
		if !strings.Contains(body, want) {
			t.Errorf("email missing %q:\n%s", want, body)
		}
	}
}
