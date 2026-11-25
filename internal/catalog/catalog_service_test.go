package catalog

import (
	"testing"
	"time"

	"safari-quest-api/internal/fileresource"
	"safari-quest-api/internal/touroperator"
	"safari-quest-api/internal/tourpackage"
)

func pkg(start time.Time, days, pax int, files ...string) tourpackage.Package {
	p := tourpackage.Package{StartDate: start, EndDate: start.AddDate(0, 0, days-1), Pax: pax,
		Operator: touroperator.TourOperator{IsActive: true, IsVerified: true}}
	for _, f := range files {
		p.FileResources = append(p.FileResources, fileresource.FileResource{Path: f})
	}
	return p
}

func TestImagePathPicksFirstImage(t *testing.T) {
	p := pkg(time.Now(), 1, 1, "uploads/brochure.pdf", "uploads/a.JPG", "uploads/b.png")
	if got := imagePath(p); got != "uploads/a.JPG" {
		t.Fatalf("imagePath = %q, want the first image", got)
	}
	if got := imagePath(pkg(time.Now(), 1, 1, "uploads/brochure.pdf")); got != "" {
		t.Fatalf("imagePath without images = %q, want empty", got)
	}
}

func TestToPublicSeatsAndBookable(t *testing.T) {
	future := time.Now().AddDate(0, 0, 10)
	p := toPublic(pkg(future, 7, 8), tourpackage.Rating{}, 3)
	if p.SeatsLeft != 5 || !p.Bookable || p.DurationDays != 7 {
		t.Fatalf("future package: seats %d bookable %v days %d", p.SeatsLeft, p.Bookable, p.DurationDays)
	}
	full := toPublic(pkg(future, 2, 4), tourpackage.Rating{}, 6)
	if full.SeatsLeft != 0 || full.Bookable {
		t.Fatalf("overbooked package should show 0 seats and not be bookable: %+v", full)
	}
	past := toPublic(pkg(time.Now().AddDate(0, 0, -3), 5, 4), tourpackage.Rating{}, 0)
	if past.Bookable {
		t.Fatal("a package that already started must not be bookable")
	}
}

func TestUnverifiedCompanyIsNotBookable(t *testing.T) {
	p := pkg(time.Now().AddDate(0, 0, 10), 3, 4)
	p.Operator.IsVerified = false
	if toPublic(p, tourpackage.Rating{}, 0).Bookable {
		t.Fatal("a package from an unverified company must not be bookable")
	}
}
