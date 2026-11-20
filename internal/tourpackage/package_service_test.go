package tourpackage

import "testing"

func TestParseTrip(t *testing.T) {
	input := Input{StartDate: "2026-12-01", EndDate: "2026-12-05", Pax: 4, Price: 1200}
	if _, _, err := parseTrip(input); err != nil {
		t.Fatal(err)
	}
	input.EndDate = "2026-11-01"
	if _, _, err := parseTrip(input); err != ErrInvalidPackage {
		t.Fatalf("end before start: %v", err)
	}
	input.EndDate = "2026-12-05"
	input.Pax = 0
	if _, _, err := parseTrip(input); err != ErrInvalidPackage {
		t.Fatalf("pax: %v", err)
	}
	input.Pax = 2
	input.StartDate = "December"
	if _, _, err := parseTrip(input); err != ErrInvalidPackage {
		t.Fatalf("date: %v", err)
	}
}
