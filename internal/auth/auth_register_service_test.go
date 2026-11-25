package auth

import (
	"strings"
	"testing"
)

func TestStrongPassword(t *testing.T) {
	for pw, want := range map[string]bool{
		"short1":       false,
		"lettersonly":  false,
		"12345678":     false,
		"safari2026":   true,
		"Kilimanjaro9": true,
	} {
		if got := strongPassword(pw); got != want {
			t.Errorf("strongPassword(%q) = %v, want %v", pw, got, want)
		}
	}
}

func TestSplitName(t *testing.T) {
	cases := map[string][2]string{
		"Juma Ali Mwangi": {"Juma Ali", "Mwangi"},
		"  Neema  ":       {"Neema", ""},
		"":                {"", ""},
	}
	for in, want := range cases {
		if f, l := splitName(in); f != want[0] || l != want[1] {
			t.Errorf("splitName(%q) = %q, %q; want %q, %q", in, f, l, want[0], want[1])
		}
	}
}

func TestCompanyCode(t *testing.T) {
	code := companyCode("Serengeti Trails & Co.")
	if !strings.HasPrefix(code, "SERENGETI_TRAILS_CO_") || len(code) != len("SERENGETI_TRAILS_CO_")+4 {
		t.Fatalf("companyCode = %q", code)
	}
	if a, b := companyCode("X"), companyCode("X"); a == b {
		t.Fatalf("two codes for the same name should differ: %q", a)
	}
	if code := companyCode("!!!"); !strings.HasPrefix(code, "COMPANY_") {
		t.Fatalf("empty base should fall back to COMPANY_: %q", code)
	}
}
