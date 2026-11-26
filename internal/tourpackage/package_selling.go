package tourpackage

import (
	"errors"
	"strings"
)

// How a package is sold. Several groups (Units) run on the same dates, each
// with Pax seats. Shared bookings buy seats in a group; a private booking
// takes a whole group.
const (
	ModeShared  = "shared"
	ModePrivate = "private"
	ModeBoth    = "both"

	PerPerson = "per_person" // price × travellers, charged for at least the minimum
	PerGroup  = "per_group"  // one price for the booking

	MaxUnits = 50
)

var ErrInvalidSelling = errors.New("Check how the package is sold: 1 to 50 groups, a booking mode, a pricing method for each kind of booking, a private group price when private bookings are offered, minimum travellers no larger than the seats per group, and a private group limit no larger than the number of groups")

// Selling holds the package columns that decide how it is booked and priced.
// Price (shared price) and Pax (seats per group) stay on Package.
type Selling struct {
	Units                int      `gorm:"not null;default:1"`
	BookingMode          string   `gorm:"type:varchar(10);not null;default:shared"`
	SharedPricing        string   `gorm:"type:varchar(12);not null;default:per_person"`
	SharedMinTravellers  int      `gorm:"not null;default:1"`
	PrivatePricing       *string  `gorm:"type:varchar(12)"`
	PrivatePrice         *float64 `gorm:"type:numeric(12,2)"`
	PrivateMinTravellers int      `gorm:"not null;default:1"`
	MaxPrivateUnits      *int
}

// OffersShared / OffersPrivate report which kinds of booking the package takes.
func (s Selling) OffersShared() bool  { return s.BookingMode != ModePrivate }
func (s Selling) OffersPrivate() bool { return s.BookingMode != ModeShared }

// SellingInput is the "How it's sold" part of the package form. Every field is
// optional; leaving them out keeps the old behaviour (one shared group priced
// per person).
type SellingInput struct {
	Units                int      `json:"units"`
	BookingMode          string   `json:"booking_mode"`
	SharedPricing        string   `json:"shared_pricing"`
	SharedMinTravellers  int      `json:"shared_min_travellers"`
	PrivatePricing       string   `json:"private_pricing"`
	PrivatePrice         *float64 `json:"private_price"`
	PrivateMinTravellers int      `json:"private_min_travellers"`
	MaxPrivateUnits      *int     `json:"max_private_units"`
}

// SellingResponse is shown wherever a package is shown.
type SellingResponse struct {
	Units                int      `json:"units"`
	BookingMode          string   `json:"booking_mode"`
	SharedPricing        string   `json:"shared_pricing"`
	SharedMinTravellers  int      `json:"shared_min_travellers"`
	PrivatePricing       *string  `json:"private_pricing"`
	PrivatePrice         *float64 `json:"private_price"`
	PrivateMinTravellers int      `json:"private_min_travellers"`
	MaxPrivateUnits      *int     `json:"max_private_units"`
}

func SellingOf(p Package) SellingResponse {
	s := p.Selling
	return SellingResponse{
		Units:                s.Units,
		BookingMode:          s.BookingMode,
		SharedPricing:        s.SharedPricing,
		SharedMinTravellers:  s.SharedMinTravellers,
		PrivatePricing:       s.PrivatePricing,
		PrivatePrice:         s.PrivatePrice,
		PrivateMinTravellers: s.PrivateMinTravellers,
		MaxPrivateUnits:      s.MaxPrivateUnits,
	}
}

func orDefault(v, fallback int) int {
	if v == 0 {
		return fallback
	}
	return v
}

func pricingOrDefault(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return PerPerson
	}
	return v
}

// sellingFrom validates the form against the seats per group and returns the
// columns to store. Settings for a kind of booking the package does not offer
// are reset, so they cannot apply by accident.
func sellingFrom(in SellingInput, seatsPerGroup int) (Selling, error) {
	s := Selling{
		Units:                orDefault(in.Units, 1),
		BookingMode:          strings.ToLower(strings.TrimSpace(in.BookingMode)),
		SharedPricing:        pricingOrDefault(in.SharedPricing),
		SharedMinTravellers:  orDefault(in.SharedMinTravellers, 1),
		PrivateMinTravellers: orDefault(in.PrivateMinTravellers, 1),
		MaxPrivateUnits:      in.MaxPrivateUnits,
	}
	if s.BookingMode == "" {
		s.BookingMode = ModeShared
	}
	validPricing := func(p string) bool { return p == PerPerson || p == PerGroup }
	switch {
	case s.Units < 1 || s.Units > MaxUnits,
		s.BookingMode != ModeShared && s.BookingMode != ModePrivate && s.BookingMode != ModeBoth,
		!validPricing(s.SharedPricing),
		s.SharedMinTravellers < 1 || s.SharedMinTravellers > seatsPerGroup:
		return Selling{}, ErrInvalidSelling
	}

	if s.OffersPrivate() {
		pricing := pricingOrDefault(in.PrivatePricing)
		if !validPricing(pricing) || in.PrivatePrice == nil || *in.PrivatePrice < 0 ||
			s.PrivateMinTravellers < 1 || s.PrivateMinTravellers > seatsPerGroup {
			return Selling{}, ErrInvalidSelling
		}
		price := *in.PrivatePrice
		s.PrivatePricing, s.PrivatePrice = &pricing, &price
		if s.MaxPrivateUnits != nil && (*s.MaxPrivateUnits < 1 || *s.MaxPrivateUnits > s.Units) {
			return Selling{}, ErrInvalidSelling
		}
		// Private-only: every group can go private, so a cap means nothing.
		if s.BookingMode == ModePrivate {
			s.MaxPrivateUnits = nil
		}
	} else {
		s.PrivatePricing, s.PrivatePrice, s.PrivateMinTravellers, s.MaxPrivateUnits = nil, nil, 1, nil
	}
	if !s.OffersShared() {
		s.SharedPricing, s.SharedMinTravellers = PerPerson, 1
	}
	// A minimum only changes the price when charging per person.
	if s.SharedPricing == PerGroup {
		s.SharedMinTravellers = 1
	}
	if s.PrivatePricing != nil && *s.PrivatePricing == PerGroup {
		s.PrivateMinTravellers = 1
	}
	return s, nil
}

var ErrBookedBeyond = errors.New("Existing bookings don't fit these settings: keep enough groups and seats per group for them, and keep offering the kinds of booking (shared or private) they were made as")

// groupUse is what active bookings hold in one group of a package.
type groupUse struct {
	UnitNumber int
	Pax        int
	Private    bool
	Shared     bool
}

// fitsBookings refuses package changes that would strand active bookings:
// fewer groups or seats than booked, or no longer offering a booking type
// that existing bookings use. Bookings are read with SQL so this package does
// not depend on the booking feature.
func fitsBookings(p Package) error {
	uses, err := activeGroupUse(p.ID)
	if err != nil {
		return err
	}
	for _, u := range uses {
		if u.UnitNumber > p.Selling.Units || u.Pax > p.Pax ||
			(u.Private && !p.Selling.OffersPrivate()) || (u.Shared && !p.Selling.OffersShared()) {
			return ErrBookedBeyond
		}
	}
	return nil
}
