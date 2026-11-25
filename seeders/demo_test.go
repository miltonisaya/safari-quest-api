package seeders

import (
	"encoding/json"
	"testing"
)

func TestDemoPackagesAreValid(t *testing.T) {
	var items []demoPackage
	if err := json.Unmarshal(demoPackagesJSON, &items); err != nil {
		t.Fatal(err)
	}
	if len(items) == 0 {
		t.Fatal("no demo packages")
	}
	meals := map[string]bool{"breakfast": true, "lunch": true, "dinner": true}
	for _, p := range items {
		if p.Name == "" || p.Operator == "" || p.Location == "" || p.PackageType == "" || p.Pax < 1 || p.Price < 0 {
			t.Errorf("%q is missing required fields", p.Name)
		}
		if len(p.Itinerary) > p.Days {
			t.Errorf("%q has %d itinerary days for a %d-day trip", p.Name, len(p.Itinerary), p.Days)
		}
		if p.Difficulty != nil && *p.Difficulty != "easy" && *p.Difficulty != "moderate" && *p.Difficulty != "challenging" {
			t.Errorf("%q has difficulty %q", p.Name, *p.Difficulty)
		}
		for _, d := range p.Itinerary {
			for _, m := range d.Meals {
				if !meals[m] {
					t.Errorf("%q day %q has meal %q", p.Name, d.Title, m)
				}
			}
		}
	}
}

func TestDemoCode(t *testing.T) {
	if got := demoCode("Baobab Safari Co."); got != "DEMO_BAOBAB_SAFARI_CO" {
		t.Fatalf("demoCode = %q", got)
	}
}
