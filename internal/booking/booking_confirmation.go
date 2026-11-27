package booking

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"safari-quest-api/internal/companyscope"
	"safari-quest-api/internal/setting"
	"safari-quest-api/internal/tourpackage"
	"safari-quest-api/internal/user"
	ws "safari-quest-api/websocket"

	"github.com/google/uuid"
)

var (
	ErrNotPending    = errors.New("Only pending bookings can be confirmed or declined")
	ErrDeclineReason = errors.New("Tell the traveler why: a reason of 3 to 500 characters")
)

// ExpiryCheckEvery is how often overdue pending bookings are expired.
const ExpiryCheckEvery = 5 * time.Minute

// confirmDeadline is when a booking made at `from` expires if still pending;
// nil when the administrator set no limit (0 hours).
func confirmDeadline(from time.Time) *time.Time {
	hours := setting.Int(setting.BookingConfirmHours)
	if hours <= 0 {
		return nil
	}
	t := from.Add(time.Duration(hours) * time.Hour)
	return &t
}

// pendingFor loads a booking the staff member may act on and checks it is
// still waiting for an answer.
func pendingFor(id uuid.UUID, scope companyscope.Scope) (Booking, error) {
	b, err := FindByUUID(id)
	if err != nil {
		return Booking{}, err
	}
	if !scope.Allows(b.Package.OperatorID) {
		return Booking{}, ErrOutsideCompany
	}
	if b.Status != StatusPending {
		return Booking{}, ErrNotPending
	}
	return b, nil
}

func saveAndAnnounce(b *Booking, byTraveler bool) (Response, error) {
	if err := Save(b, mailsStatus(byTraveler)); err != nil {
		return Response{}, err
	}
	saved, err := FindByUUID(b.UUID)
	if err != nil {
		return Response{}, err
	}
	resp := ToResponse(saved)
	publish(ws.ActionUpdated, resp)
	return resp, nil
}

// Confirm accepts a pending booking. Its seats stay held and the traveler is told.
func Confirm(id uuid.UUID, scope companyscope.Scope) (Response, error) {
	b, err := pendingFor(id, scope)
	if err != nil {
		return Response{}, err
	}
	if started(b.Package) {
		return Response{}, ErrDeparted
	}
	if err := applyStatus(&b, StatusConfirmed); err != nil {
		return Response{}, err
	}
	b.ExpiresAt = nil
	return saveAndAnnounce(&b, false)
}

// DeclineInput says why the company turned the request down; the traveler sees it.
type DeclineInput struct {
	Reason string `json:"reason" binding:"required"`
}

// Decline turns a pending booking down with a reason and frees its seats.
func Decline(id uuid.UUID, input DeclineInput, scope companyscope.Scope) (Response, error) {
	reason := strings.TrimSpace(input.Reason)
	if len(reason) < 3 || len(reason) > 500 {
		return Response{}, ErrDeclineReason
	}
	b, err := pendingFor(id, scope)
	if err != nil {
		return Response{}, err
	}
	if err := applyStatus(&b, StatusDeclined); err != nil {
		return Response{}, err
	}
	b.DeclineReason = &reason
	b.ExpiresAt = nil
	return saveAndAnnounce(&b, false)
}

// ExpireOverdue expires pending bookings whose confirm-by time has passed,
// freeing their seats and telling the traveler and the company.
func ExpireOverdue(now time.Time) (int, error) {
	overdue, err := findOverdue(now)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, b := range overdue {
		if err := applyStatus(&b, StatusExpired); err != nil {
			continue
		}
		if _, err := saveAndAnnounce(&b, false); err != nil {
			log.Printf("[booking] could not expire %s: %v\n", b.Code, err)
			continue
		}
		n++
	}
	return n, nil
}

// RunExpiry expires overdue bookings now and then every ExpiryCheckEvery
// until ctx ends. Started once from main.
func RunExpiry(ctx context.Context) {
	run := func() {
		if n, err := ExpireOverdue(time.Now()); err != nil {
			log.Printf("[booking] expiry check failed: %v\n", err)
		} else if n > 0 {
			log.Printf("[booking] expired %d unconfirmed booking(s)\n", n)
		}
	}
	run()
	ticker := time.NewTicker(ExpiryCheckEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}

// ── Traveler's booking detail (confirmation voucher) ─────────────────────────

// CompanyContact is how the traveler reaches the tour company.
type CompanyContact struct {
	Name    string  `json:"name"`
	Email   string  `json:"email"`
	Phone   string  `json:"phone"`
	Website *string `json:"website"`
}

// Detail is one of the traveler's bookings with everything a confirmation
// voucher needs.
type Detail struct {
	Response
	Trip               tourpackage.Response `json:"trip"`
	Company            CompanyContact       `json:"company"`
	PreferredLanguages []string             `json:"preferred_languages"`
	Travellers         []TravellerResponse  `json:"travellers"` // names, age groups, nationalities; no ID numbers
}

// MyBooking returns one of the traveler's own bookings in full.
func MyBooking(travelerUUID, bookingUUID uuid.UUID) (Detail, error) {
	traveler, err := user.FindByUUID(travelerUUID)
	if err != nil {
		return Detail{}, err
	}
	b, err := FindByUUID(bookingUUID)
	if err != nil {
		return Detail{}, err
	}
	if b.UserID != traveler.ID {
		return Detail{}, ErrNotYours
	}
	pkg, err := tourpackage.FindByUUID(b.Package.UUID)
	if err != nil {
		return Detail{}, err
	}
	trip := tourpackage.ToResponse(pkg)
	trip.FileResources = []tourpackage.FileRef{}
	details, err := detailsOf(b)
	if err != nil {
		return Detail{}, err
	}
	for i := range details.Travellers {
		details.Travellers[i].IDNumber, details.Travellers[i].IDDocument = nil, nil
	}
	return Detail{
		Response:           ToResponse(b),
		Trip:               trip,
		PreferredLanguages: details.PreferredLanguages,
		Travellers:         details.Travellers,
		Company: CompanyContact{
			Name: pkg.Operator.Name, Email: pkg.Operator.Email, Phone: pkg.Operator.Phone, Website: pkg.Operator.Website,
		},
	}, nil
}
