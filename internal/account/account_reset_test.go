package account

import (
	"strings"
	"testing"

	"safari-quest-api/config"
	"safari-quest-api/internal/user"
)

func TestCodesAreSixDigits(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		c, err := newCode()
		if err != nil || len(c) != 6 || strings.Trim(c, "0123456789") != "" {
			t.Fatalf("bad code %q (%v)", c, err)
		}
		seen[c] = true
	}
	if len(seen) < 45 {
		t.Fatalf("codes repeat too often: %d unique of 50", len(seen))
	}
}

func TestCodeHashIsBoundToUserAndSecret(t *testing.T) {
	config.App.JWTSecret = "test-secret"
	a, b := hashCode(1, "123456"), hashCode(2, "123456")
	if len(a) != 64 || a == b || a != hashCode(1, "123456") {
		t.Fatalf("hash must be stable per user and differ between users: %s %s", a, b)
	}
}

func TestCodeEmail(t *testing.T) {
	d := codeEmail(user.User{FirstName: "Asha", Email: "asha@example.test"}, "042731")
	if d.To != "asha@example.test" || d.Content.Code != "042731" || !strings.Contains(d.Content.Subject, "042731") {
		t.Fatalf("email = %+v", d)
	}
	if !strings.Contains(d.Content.Text(), "10 minutes") {
		t.Fatalf("email should say how long the code works:\n%s", d.Content.Text())
	}
}
