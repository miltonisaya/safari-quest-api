package booking

import (
	"errors"
	"fmt"
	"strings"

	"safari-quest-api/internal/tourpackage"
)

// Booking types: shared seats in a group, or a whole group for one party.
const (
	TypeShared  = "shared"
	TypePrivate = "private"
)

var (
	ErrTypeNotOffered = errors.New("This trip isn't offered for that kind of booking")
	ErrNoPrivateGroup = errors.New("No private groups are left on this trip")
)

// errGroupTooSmall says how many travellers fit in one group.
func errGroupTooSmall(seats int) error {
	return fmt.Errorf("%w: a group on this trip has %d seats", ErrNotEnoughSeats, seats)
}

// bookingType normalises the requested type; empty means shared.
func bookingType(t string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "", TypeShared:
		return TypeShared, nil
	case TypePrivate:
		return TypePrivate, nil
	default:
		return "", ErrTypeNotOffered
	}
}

// Quote is a booking's price, kept on the booking so later price changes
// never alter it.
type Quote struct {
	Pricing           string  // per_person or per_group
	UnitPrice         float64 // price per person, or for the whole booking
	MinTravellers     int     // per person: charged for at least this many
	ChargedTravellers int
	Total             float64
}

// quote prices pax travellers for the given booking type of the package.
func quote(p tourpackage.Package, kind string, pax int) (Quote, error) {
	s := p.Selling
	var q Quote
	switch kind {
	case TypeShared:
		if !s.OffersShared() {
			return Quote{}, ErrTypeNotOffered
		}
		q = Quote{Pricing: s.SharedPricing, UnitPrice: p.Price, MinTravellers: s.SharedMinTravellers}
	case TypePrivate:
		if !s.OffersPrivate() || s.PrivatePricing == nil || s.PrivatePrice == nil {
			return Quote{}, ErrTypeNotOffered
		}
		q = Quote{Pricing: *s.PrivatePricing, UnitPrice: *s.PrivatePrice, MinTravellers: s.PrivateMinTravellers}
	default:
		return Quote{}, ErrTypeNotOffered
	}
	if pax < 1 {
		return Quote{}, ErrNotEnoughSeats
	}
	if pax > p.Pax {
		return Quote{}, errGroupTooSmall(p.Pax)
	}
	return reprice(q, pax), nil
}

// reprice recalculates a quote for a new number of travellers.
func reprice(q Quote, pax int) Quote {
	if q.MinTravellers < 1 {
		q.MinTravellers = 1
	}
	q.ChargedTravellers = pax
	if q.Pricing == tourpackage.PerGroup {
		q.Total = q.UnitPrice
		return q
	}
	if q.ChargedTravellers < q.MinTravellers {
		q.ChargedTravellers = q.MinTravellers
	}
	q.Total = q.UnitPrice * float64(q.ChargedTravellers)
	return q
}

// groupUse is what active bookings hold in one group.
type groupUse struct {
	Held    int  // travellers in the group
	Private bool // a private booking owns the group
}

// usage maps a group number (1…Units) to what is held in it.
type usage map[int]groupUse

func (u usage) privateGroups() int {
	n := 0
	for _, g := range u {
		if g.Private {
			n++
		}
	}
	return n
}

// privateRoomLeft is how many more groups may go private under the cap.
func privateRoomLeft(p tourpackage.Package, u usage) int {
	left := p.Selling.Units
	if p.Selling.MaxPrivateUnits != nil {
		left = *p.Selling.MaxPrivateUnits
	}
	return left - u.privateGroups()
}

// assignGroup picks the group for a new booking.
//
// Shared bookings go to the partly filled group they fit most tightly, and
// only then to an empty group, so whole groups stay free for private parties.
// A booking is never split across groups. Private bookings take the lowest
// numbered empty group, within the operator's cap on private groups.
func assignGroup(p tourpackage.Package, kind string, pax int, u usage) (int, error) {
	if kind == TypePrivate {
		if privateRoomLeft(p, u) <= 0 {
			return 0, ErrNoPrivateGroup
		}
		for n := 1; n <= p.Selling.Units; n++ {
			if g := u[n]; g.Held == 0 && !g.Private {
				return n, nil
			}
		}
		return 0, ErrNoPrivateGroup
	}

	best, bestFree, firstEmpty := 0, 0, 0
	for n := 1; n <= p.Selling.Units; n++ {
		g := u[n]
		free := p.Pax - g.Held
		if g.Private || free < pax {
			continue
		}
		if g.Held == 0 {
			if firstEmpty == 0 {
				firstEmpty = n
			}
			continue
		}
		if best == 0 || free < bestFree {
			best, bestFree = n, free
		}
	}
	if best != 0 {
		return best, nil
	}
	if firstEmpty != 0 {
		return firstEmpty, nil
	}
	return 0, ErrNotEnoughSeats
}

// fitsGroup reports whether pax travellers of the given type fit in group n,
// given what other bookings hold there (used when a booking is edited).
func fitsGroup(p tourpackage.Package, kind string, n, pax int, others usage) bool {
	g := others[n]
	if n < 1 || n > p.Selling.Units || pax > p.Pax {
		return false
	}
	if kind == TypePrivate {
		return g.Held == 0 && !g.Private
	}
	return !g.Private && g.Held+pax <= p.Pax
}

// Availability is what is still on sale for a package.
type Availability struct {
	SeatsLeft           int `json:"seats_left"`            // shared seats across all groups
	MaxSharedTravellers int `json:"max_shared_travellers"` // most travellers one shared booking can bring (one group)
	PrivateGroupsLeft   int `json:"private_groups_left"`
}

// Bookable reports whether anything is still on sale.
func (a Availability) Bookable() bool { return a.SeatsLeft > 0 || a.PrivateGroupsLeft > 0 }

func availabilityOf(p tourpackage.Package, u usage) Availability {
	var a Availability
	empty := 0
	for n := 1; n <= p.Selling.Units; n++ {
		g := u[n]
		if g.Private {
			continue
		}
		if g.Held == 0 {
			empty++
		}
		if p.Selling.OffersShared() {
			free := p.Pax - g.Held
			if free > 0 {
				a.SeatsLeft += free
			}
			if free > a.MaxSharedTravellers {
				a.MaxSharedTravellers = free
			}
		}
	}
	if p.Selling.OffersPrivate() {
		a.PrivateGroupsLeft = min(empty, max(privateRoomLeft(p, u), 0))
	}
	return a
}
