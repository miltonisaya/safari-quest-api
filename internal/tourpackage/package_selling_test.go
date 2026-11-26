package tourpackage

import (
	"errors"
	"testing"
)

func ptrF(v float64) *float64 { return &v }
func ptrI(v int) *int         { return &v }

func TestSellingDefaultsKeepOldBehaviour(t *testing.T) {
	s, err := sellingFrom(SellingInput{}, 8)
	if err != nil {
		t.Fatal(err)
	}
	if s.Units != 1 || s.BookingMode != ModeShared || s.SharedPricing != PerPerson || s.SharedMinTravellers != 1 ||
		s.PrivatePricing != nil || s.PrivatePrice != nil {
		t.Fatalf("defaults = %+v", s)
	}
}

func TestSellingBothModes(t *testing.T) {
	s, err := sellingFrom(SellingInput{
		Units: 3, BookingMode: "Both", SharedPricing: "per_person", SharedMinTravellers: 2,
		PrivatePricing: "per_group", PrivatePrice: ptrF(9000), PrivateMinTravellers: 4, MaxPrivateUnits: ptrI(2),
	}, 6)
	if err != nil {
		t.Fatal(err)
	}
	if s.BookingMode != ModeBoth || s.SharedMinTravellers != 2 || *s.PrivatePricing != PerGroup || *s.PrivatePrice != 9000 {
		t.Fatalf("selling = %+v", s)
	}
	if s.PrivateMinTravellers != 1 {
		t.Fatalf("per-group pricing should drop the minimum, got %d", s.PrivateMinTravellers)
	}
	if *s.MaxPrivateUnits != 2 {
		t.Fatalf("max private units = %d", *s.MaxPrivateUnits)
	}
}

func TestSellingResetsUnusedSide(t *testing.T) {
	s, err := sellingFrom(SellingInput{BookingMode: "shared", PrivatePricing: "per_group", PrivatePrice: ptrF(1), MaxPrivateUnits: ptrI(1)}, 4)
	if err != nil || s.PrivatePrice != nil || s.PrivatePricing != nil || s.MaxPrivateUnits != nil {
		t.Fatalf("shared-only should drop private settings: %+v %v", s, err)
	}
	s, err = sellingFrom(SellingInput{BookingMode: "private", Units: 2, SharedMinTravellers: 3, PrivatePrice: ptrF(500), MaxPrivateUnits: ptrI(1)}, 4)
	if err != nil || s.SharedMinTravellers != 1 || s.MaxPrivateUnits != nil || *s.PrivatePricing != PerPerson {
		t.Fatalf("private-only should drop shared settings and the cap: %+v %v", s, err)
	}
}

func TestSellingRejectsBadInput(t *testing.T) {
	bad := map[string]SellingInput{
		"too many groups":            {Units: 51},
		"unknown mode":               {BookingMode: "vip"},
		"unknown pricing":            {SharedPricing: "per_day"},
		"minimum over seats":         {SharedMinTravellers: 9},
		"private without price":      {BookingMode: "both"},
		"negative private price":     {BookingMode: "both", PrivatePrice: ptrF(-1)},
		"private minimum over seats": {BookingMode: "private", PrivatePrice: ptrF(1), PrivateMinTravellers: 9},
		"cap over groups":            {BookingMode: "both", Units: 2, PrivatePrice: ptrF(1), MaxPrivateUnits: ptrI(3)},
	}
	for name, in := range bad {
		if _, err := sellingFrom(in, 8); !errors.Is(err, ErrInvalidSelling) {
			t.Errorf("%s: got %v, want ErrInvalidSelling", name, err)
		}
	}
}
