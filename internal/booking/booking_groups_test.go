package booking

import (
	"errors"
	"testing"

	"safari-quest-api/internal/tourpackage"
)

func str(s string) *string   { return &s }
func num(v float64) *float64 { return &v }
func intp(v int) *int        { return &v }

// trip: 3 groups of 6 seats, shared $1,000 pp (min 2), private $5,000 per group.
func trip() tourpackage.Package {
	return tourpackage.Package{
		Pax:   6,
		Price: 1000,
		Selling: tourpackage.Selling{
			Units: 3, BookingMode: tourpackage.ModeBoth,
			SharedPricing: tourpackage.PerPerson, SharedMinTravellers: 2,
			PrivatePricing: str(tourpackage.PerGroup), PrivatePrice: num(5000), PrivateMinTravellers: 1,
		},
	}
}

func TestQuote(t *testing.T) {
	p := trip()
	solo, _ := quote(p, TypeShared, 1)
	if solo.ChargedTravellers != 2 || solo.Total != 2000 {
		t.Fatalf("solo traveller should pay the minimum of 2: %+v", solo)
	}
	three, _ := quote(p, TypeShared, 3)
	if three.Total != 3000 {
		t.Fatalf("3 travellers = %v", three.Total)
	}
	private, _ := quote(p, TypePrivate, 4)
	if private.Pricing != tourpackage.PerGroup || private.Total != 5000 || private.ChargedTravellers != 4 {
		t.Fatalf("private group = %+v", private)
	}

	p.Selling.PrivatePricing, p.Selling.PrivatePrice, p.Selling.PrivateMinTravellers = str(tourpackage.PerPerson), num(1200), 4
	small, _ := quote(p, TypePrivate, 3)
	if small.ChargedTravellers != 4 || small.Total != 4800 {
		t.Fatalf("private per person should charge the minimum of 4: %+v", small)
	}

	if _, err := quote(p, TypeShared, 7); !errors.Is(err, ErrNotEnoughSeats) {
		t.Fatalf("7 travellers in 6-seat groups: %v", err)
	}
	p.Selling.BookingMode = tourpackage.ModeShared
	if _, err := quote(p, TypePrivate, 2); !errors.Is(err, ErrTypeNotOffered) {
		t.Fatalf("private on a shared-only trip: %v", err)
	}
}

func TestSharedFillsPartlyBookedGroupsFirst(t *testing.T) {
	p := trip()
	u := usage{2: {Held: 4}}
	if n, _ := assignGroup(p, TypeShared, 2, u); n != 2 {
		t.Fatalf("2 travellers should join group 2 (2 seats free), got %d", n)
	}
	if n, _ := assignGroup(p, TypeShared, 3, u); n != 1 {
		t.Fatalf("3 travellers don't fit group 2, should start group 1, got %d", n)
	}
	u = usage{1: {Held: 5}, 2: {Held: 3}}
	if n, _ := assignGroup(p, TypeShared, 1, u); n != 1 {
		t.Fatalf("tightest fit for 1 traveller is group 1, got %d", n)
	}
}

func TestSharedNeverSplitsOrEntersPrivateGroups(t *testing.T) {
	p := trip()
	u := usage{1: {Held: 4}, 2: {Held: 4}, 3: {Held: 1, Private: true}}
	if _, err := assignGroup(p, TypeShared, 3, u); !errors.Is(err, ErrNotEnoughSeats) {
		t.Fatalf("4 seats free in total but at most 2 together: %v", err)
	}
}

func TestPrivateTakesEmptyGroupWithinCap(t *testing.T) {
	p := trip()
	u := usage{1: {Held: 2}}
	if n, _ := assignGroup(p, TypePrivate, 4, u); n != 2 {
		t.Fatalf("private should take empty group 2, got %d", n)
	}
	p.Selling.MaxPrivateUnits = intp(1)
	u = usage{1: {Held: 3, Private: true}}
	if _, err := assignGroup(p, TypePrivate, 2, u); !errors.Is(err, ErrNoPrivateGroup) {
		t.Fatalf("cap of 1 private group reached: %v", err)
	}
	u = usage{1: {Held: 1}, 2: {Held: 1}, 3: {Held: 1}}
	p.Selling.MaxPrivateUnits = nil
	if _, err := assignGroup(p, TypePrivate, 2, u); !errors.Is(err, ErrNoPrivateGroup) {
		t.Fatalf("no empty group left: %v", err)
	}
}

func TestAvailability(t *testing.T) {
	p := trip()
	a := availabilityOf(p, usage{1: {Held: 6, Private: true}, 2: {Held: 4}})
	if a.SeatsLeft != 8 || a.MaxSharedTravellers != 6 || a.PrivateGroupsLeft != 1 || !a.Bookable() {
		t.Fatalf("availability = %+v", a)
	}
	p.Selling.MaxPrivateUnits = intp(1)
	if a := availabilityOf(p, usage{1: {Held: 2, Private: true}}); a.PrivateGroupsLeft != 0 {
		t.Fatalf("cap reached, private groups left = %d", a.PrivateGroupsLeft)
	}
	full := availabilityOf(p, usage{1: {Held: 6}, 2: {Held: 6}, 3: {Held: 6}})
	if full.Bookable() {
		t.Fatalf("full trip should not be bookable: %+v", full)
	}
}

func TestFitsGroupWhenEditing(t *testing.T) {
	p := trip()
	others := usage{1: {Held: 3}}
	if !fitsGroup(p, TypeShared, 1, 3, others) || fitsGroup(p, TypeShared, 1, 4, others) {
		t.Fatal("shared edit: 3 more fit in group 1, 4 do not")
	}
	if fitsGroup(p, TypePrivate, 1, 2, others) || !fitsGroup(p, TypePrivate, 2, 6, others) {
		t.Fatal("private edit needs the group to itself")
	}
}
