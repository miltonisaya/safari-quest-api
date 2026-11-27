package booking

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"safari-quest-api/internal/companyscope"
	"safari-quest-api/internal/tourpackage"
	"safari-quest-api/internal/user"
	"safari-quest-api/pkg/pagination"
	ws "safari-quest-api/websocket"

	"github.com/google/uuid"
)

var (
	ErrOutsideCompany   = errors.New("You can only manage bookings in your company")
	ErrInvalidReference = errors.New("invalid reference")
	ErrDeparted         = errors.New("This package has already started and can no longer be booked or changed")
	ErrInvalidStatus    = errors.New("That status change is not allowed")
	ErrFinal            = errors.New("Completed and cancelled bookings can no longer be changed")
	ErrNotYours         = errors.New("You can only manage your own bookings")
	ErrInactiveTraveler = errors.New("This traveler's account is inactive")
	ErrUnavailable      = errors.New("This trip isn't open for booking yet")
)

// StaffInput is what company staff send to record a booking for a traveler.
type StaffInput struct {
	PackageUUID  uuid.UUID `json:"package_uuid" binding:"required"`
	TravelerUUID uuid.UUID `json:"traveler_uuid" binding:"required"`
	BookingType  string    `json:"booking_type"` // shared (default) or private
	Pax          int       `json:"pax" binding:"required,min=1"`
	Notes        *string   `json:"notes"`
	Status       Status    `json:"status"` // pending (default) or confirmed
}

// UpdateInput changes guests, notes or status. Status follows CanMoveTo.
type UpdateInput struct {
	Pax    int     `json:"pax" binding:"required,min=1"`
	Notes  *string `json:"notes"`
	Status Status  `json:"status" binding:"required"`
}

// TravelerInput is what a signed-in traveler sends to book a package.
type TravelerInput struct {
	PackageUUID uuid.UUID `json:"package_uuid" binding:"required"`
	BookingType string    `json:"booking_type"` // shared (default) or private
	Pax         int       `json:"pax" binding:"required,min=1"`
	Notes       *string   `json:"notes"`
}

type Ref struct {
	UUID uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
	Code string    `json:"code"`
}

type PackageRef struct {
	UUID        uuid.UUID `json:"uuid"`
	Name        string    `json:"name"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
	PackageType string    `json:"package_type"`
}

type TravelerRef struct {
	UUID   uuid.UUID `json:"uuid"`
	Name   string    `json:"name"`
	Email  string    `json:"email"`
	Mobile string    `json:"mobile"`
}

type Response struct {
	UUID               uuid.UUID   `json:"uuid"`
	Code               string      `json:"code"`
	Package            PackageRef  `json:"package"`
	TourOperator       Ref         `json:"tour_operator"`
	Traveler           TravelerRef `json:"traveler"`
	Pax                int         `json:"pax"`
	BookingType        string      `json:"booking_type"`
	Group              int         `json:"group"`
	Pricing            string      `json:"pricing"`
	MinTravellers      int         `json:"min_travellers"`
	ChargedTravellers  int         `json:"charged_travellers"`
	UnitPrice          float64     `json:"unit_price"`
	TotalPrice         float64     `json:"total_price"`
	Status             Status      `json:"status"`
	Notes              *string     `json:"notes"`
	ConfirmedAt        *time.Time  `json:"confirmed_at"`
	CancelledAt        *time.Time  `json:"cancelled_at"`
	DeclineReason      *string     `json:"decline_reason"`
	DeclinedAt         *time.Time  `json:"declined_at"`
	ExpiresAt          *time.Time  `json:"expires_at"` // confirm-by deadline while pending
	ExpiredAt          *time.Time  `json:"expired_at"`
	DetailsSubmittedAt *time.Time  `json:"details_submitted_at"` // traveller details given; nil = still needed
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
}

// BookableResponse is a package a traveler can still book.
type BookableResponse struct {
	UUID         uuid.UUID `json:"uuid"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	TourOperator Ref       `json:"tour_operator"`
	StartDate    string    `json:"start_date"`
	EndDate      string    `json:"end_date"`
	PackageType  string    `json:"package_type"`
	PricePerHead float64   `json:"price_per_person"`
	Pax          int       `json:"pax"`
	tourpackage.SellingResponse
	Availability
	Difficulty *string                    `json:"difficulty"`
	Rating     tourpackage.RatingResponse `json:"rating"`
}

// BookableDetail is everything a traveler sees before booking a package.
type BookableDetail struct {
	tourpackage.Response
	PricePerHead float64 `json:"price_per_person"`
	Availability
	Bookable bool `json:"bookable"` // false once started, full or not published
}

func travelerRef(u user.User) TravelerRef {
	return TravelerRef{
		UUID:   u.UUID,
		Name:   strings.TrimSpace(u.FirstName + " " + u.LastName),
		Email:  u.Email,
		Mobile: u.Mobile,
	}
}

func ToResponse(b Booking) Response {
	return Response{
		UUID: b.UUID,
		Code: b.Code,
		Package: PackageRef{
			UUID:        b.Package.UUID,
			Name:        b.Package.Name,
			StartDate:   b.Package.StartDate.Format("2006-01-02"),
			EndDate:     b.Package.EndDate.Format("2006-01-02"),
			PackageType: b.Package.PackageType,
		},
		TourOperator:       Ref{UUID: b.TourOperator.UUID, Name: b.TourOperator.Name, Code: b.TourOperator.Code},
		Traveler:           travelerRef(b.User),
		Pax:                b.Pax,
		BookingType:        b.BookingType,
		Group:              b.UnitNumber,
		Pricing:            b.Pricing,
		MinTravellers:      b.MinTravellers,
		ChargedTravellers:  b.ChargedTravellers,
		UnitPrice:          b.UnitPrice,
		TotalPrice:         b.TotalPrice,
		Status:             b.Status,
		Notes:              b.Notes,
		ConfirmedAt:        b.ConfirmedAt,
		CancelledAt:        b.CancelledAt,
		DeclineReason:      b.DeclineReason,
		DeclinedAt:         b.DeclinedAt,
		ExpiresAt:          b.ExpiresAt,
		ExpiredAt:          b.ExpiredAt,
		DetailsSubmittedAt: b.DetailsSubmittedAt,
		CreatedAt:          b.CreatedAt,
		UpdatedAt:          b.UpdatedAt,
	}
}

// ── Helpers ─────────────────────────────────────────────────────────────────

const codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no 0/O or 1/I

// newCode returns an unused booking reference such as SQ-7KQ4MD.
func newCode() (string, error) {
	for attempt := 0; attempt < 8; attempt++ {
		var sb strings.Builder
		sb.WriteString("SQ-")
		for i := 0; i < 6; i++ {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(codeAlphabet))))
			if err != nil {
				return "", err
			}
			sb.WriteByte(codeAlphabet[n.Int64()])
		}
		code := sb.String()
		taken, err := CodeExists(code)
		if err != nil {
			return "", err
		}
		if !taken {
			return code, nil
		}
	}
	return "", errors.New("could not generate a unique booking code")
}

// started reports whether the package's trip has begun.
func started(p tourpackage.Package) bool {
	today := time.Now().Format("2006-01-02")
	return p.StartDate.Format("2006-01-02") < today
}

// applyStatus moves the booking to next and stamps the matching time.
func applyStatus(b *Booking, next Status) error {
	if !ValidStatus(next) || !b.Status.CanMoveTo(next) {
		return ErrInvalidStatus
	}
	now := time.Now()
	if next == StatusConfirmed && b.ConfirmedAt == nil {
		b.ConfirmedAt = &now
	}
	if next == StatusCancelled && b.CancelledAt == nil {
		b.CancelledAt = &now
	}
	if next == StatusDeclined && b.DeclinedAt == nil {
		b.DeclinedAt = &now
	}
	if next == StatusExpired && b.ExpiredAt == nil {
		b.ExpiredAt = &now
	}
	b.Status = next
	return nil
}

// publish tells the booking's traveler, its company's staff and
// administrators about a change. Bookings hold personal data, so they are
// never broadcast to everyone connected.
func publish(action ws.Action, resp Response) {
	users, err := participants(resp.UUID)
	if err != nil {
		return
	}
	ws.GlobalHub.PublishTo(users, ws.Event{Entity: "booking", Action: action, Payload: resp})
}

// create builds and stores a new booking for traveler on the package.
func create(pkg tourpackage.Package, traveler user.User, requestedType string, pax int, notes *string, status Status) (Response, error) {
	if started(pkg) {
		return Response{}, ErrDeparted
	}
	if !traveler.IsActive {
		return Response{}, ErrInactiveTraveler
	}
	kind, err := bookingType(requestedType)
	if err != nil {
		return Response{}, err
	}
	q, err := quote(pkg, kind, pax)
	if err != nil {
		return Response{}, err
	}
	code, err := newCode()
	if err != nil {
		return Response{}, err
	}
	b := Booking{
		Code:              code,
		PackageID:         pkg.ID,
		TourOperatorID:    pkg.OperatorID,
		UserID:            traveler.ID,
		Pax:               pax,
		BookingType:       kind,
		Pricing:           q.Pricing,
		MinTravellers:     q.MinTravellers,
		ChargedTravellers: q.ChargedTravellers,
		UnitPrice:         q.UnitPrice,
		TotalPrice:        q.Total,
		Status:            StatusPending,
		Notes:             notes,
		ExpiresAt:         confirmDeadline(time.Now()),
	}
	if status != "" && status != StatusPending {
		if err := applyStatus(&b, status); err != nil {
			return Response{}, err
		}
	}
	if b.Status != StatusPending {
		b.ExpiresAt = nil // recorded as confirmed straight away
	}
	if err := Insert(&b); err != nil {
		return Response{}, err
	}
	saved, err := FindByUUID(b.UUID)
	if err != nil {
		return Response{}, err
	}
	resp := ToResponse(saved)
	publish(ws.ActionCreated, resp)
	notifyCreated(saved)
	return resp, nil
}

// ── Company staff ───────────────────────────────────────────────────────────

func GetAll(params pagination.Params, status Status, packageUUID *uuid.UUID, scope companyscope.Scope) (pagination.Result[Response], error) {
	rows, total, err := FindAll(params, Filters{
		Status: status, PackageUUID: packageUUID, Admin: scope.Admin, OperatorIDs: scope.OperatorIDs,
	})
	if err != nil {
		return pagination.Result[Response]{}, err
	}
	items := make([]Response, len(rows))
	for i, row := range rows {
		items[i] = ToResponse(row)
	}
	return pagination.Result[Response]{Items: items, Meta: pagination.NewMeta(total, params)}, nil
}

func GetByUUID(id uuid.UUID, scope companyscope.Scope) (Response, error) {
	b, err := FindByUUID(id)
	if err != nil {
		return Response{}, err
	}
	if !scope.Allows(b.Package.OperatorID) {
		return Response{}, ErrOutsideCompany
	}
	return ToResponse(b), nil
}

// Create records a booking for a traveler on one of the staff's own packages.
func Create(input StaffInput, scope companyscope.Scope) (Response, error) {
	if input.Status != "" && input.Status != StatusPending && input.Status != StatusConfirmed {
		return Response{}, ErrInvalidStatus
	}
	pkg, err := tourpackage.FindByUUID(input.PackageUUID)
	if err != nil {
		return Response{}, fmt.Errorf("%w: package_uuid", ErrInvalidReference)
	}
	if !scope.Allows(pkg.OperatorID) {
		return Response{}, ErrOutsideCompany
	}
	traveler, err := user.FindByUUID(input.TravelerUUID)
	if err != nil {
		return Response{}, fmt.Errorf("%w: traveler_uuid", ErrInvalidReference)
	}
	return create(pkg, traveler, input.BookingType, input.Pax, input.Notes, input.Status)
}

// Update changes guests, notes and status. Completed and cancelled bookings
// are final; guest changes are refused once the trip has started.
func Update(id uuid.UUID, input UpdateInput, scope companyscope.Scope) (Response, error) {
	b, err := FindByUUID(id)
	if err != nil {
		return Response{}, err
	}
	if !scope.Allows(b.Package.OperatorID) {
		return Response{}, ErrOutsideCompany
	}
	if !b.Status.HoldsSeats() {
		return Response{}, ErrFinal
	}
	if input.Pax != b.Pax {
		if started(b.Package) {
			return Response{}, ErrDeparted
		}
		if input.Pax > b.Package.Pax {
			return Response{}, errGroupTooSmall(b.Package.Pax)
		}
		// Re-price with the booking's own copy of the pricing, not today's package price.
		q := reprice(Quote{Pricing: b.Pricing, UnitPrice: b.UnitPrice, MinTravellers: b.MinTravellers}, input.Pax)
		b.Pax = input.Pax
		// The member list no longer matches: the traveler must update it.
		b.DetailsSubmittedAt = nil
		b.ChargedTravellers = q.ChargedTravellers
		b.TotalPrice = q.Total
	}
	b.Notes = input.Notes
	// Declining needs a reason and expiry is automatic: use their own actions.
	if input.Status == StatusDeclined || input.Status == StatusExpired {
		return Response{}, ErrInvalidStatus
	}
	previous := b.Status
	if err := applyStatus(&b, input.Status); err != nil {
		return Response{}, err
	}
	if b.Status != StatusPending {
		b.ExpiresAt = nil
	}
	if err := Save(&b); err != nil {
		return Response{}, err
	}
	saved, err := FindByUUID(b.UUID)
	if err != nil {
		return Response{}, err
	}
	resp := ToResponse(saved)
	publish(ws.ActionUpdated, resp)
	if saved.Status != previous {
		notifyStatus(saved, false)
	}
	return resp, nil
}

func Delete(id uuid.UUID, scope companyscope.Scope) error {
	b, err := FindByUUID(id)
	if err != nil {
		return err
	}
	if !scope.Allows(b.Package.OperatorID) {
		return ErrOutsideCompany
	}
	if err := Remove(id); err != nil {
		return err
	}
	ws.GlobalHub.Publish(ws.Event{Entity: "booking", Action: ws.ActionDeleted, Payload: map[string]string{"uuid": id.String()}})
	return nil
}

// Travelers searches traveler accounts staff can book for.
func Travelers(search string) ([]TravelerRef, error) {
	rows, err := FindTravelers(search, 20)
	if err != nil {
		return nil, err
	}
	out := make([]TravelerRef, len(rows))
	for i, u := range rows {
		out[i] = travelerRef(u)
	}
	return out, nil
}

// ── Travelers ───────────────────────────────────────────────────────────────

func Bookable(params pagination.Params) (pagination.Result[BookableResponse], error) {
	rows, total, err := FindBookable(params)
	if err != nil {
		return pagination.Result[BookableResponse]{}, err
	}
	ids := make([]uint, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}
	ratings, err := tourpackage.RatingsFor(ids)
	if err != nil {
		return pagination.Result[BookableResponse]{}, err
	}
	items := make([]BookableResponse, len(rows))
	for i, row := range rows {
		items[i] = BookableResponse{
			UUID:            row.UUID,
			Name:            row.Name,
			Description:     row.Description,
			TourOperator:    Ref{UUID: row.Operator.UUID, Name: row.Operator.Name, Code: row.Operator.Code},
			StartDate:       row.StartDate.Format("2006-01-02"),
			EndDate:         row.EndDate.Format("2006-01-02"),
			PackageType:     row.PackageType,
			PricePerHead:    row.Price,
			Pax:             row.Pax,
			SellingResponse: tourpackage.SellingOf(row.Package),
			Availability:    row.Availability,
			Difficulty:      row.Difficulty,
			Rating:          tourpackage.RatingResponse{Average: ratings[row.ID].Average, Count: ratings[row.ID].Count},
		}
	}
	return pagination.Result[BookableResponse]{Items: items, Meta: pagination.NewMeta(total, params)}, nil
}

// BookableByUUID returns a package's full details, itinerary and seats left.
// Any package can be viewed (e.g. one the traveler already took);
// Bookable says whether it can still be booked.
func BookableByUUID(id uuid.UUID) (BookableDetail, error) {
	pkg, err := tourpackage.FindByUUID(id)
	if err != nil {
		return BookableDetail{}, err
	}
	available, err := AvailabilityFor([]tourpackage.Package{pkg})
	if err != nil {
		return BookableDetail{}, err
	}
	ratings, err := tourpackage.RatingsFor([]uint{pkg.ID})
	if err != nil {
		return BookableDetail{}, err
	}
	a := available[pkg.ID]
	return BookableDetail{
		Response:     tourpackage.ToResponseRated(pkg, ratings[pkg.ID]),
		PricePerHead: pkg.Price,
		Availability: a,
		Bookable:     a.Bookable() && !started(pkg) && tourpackage.Published(pkg),
	}, nil
}

func MyBookings(params pagination.Params, status Status, travelerUUID uuid.UUID) (pagination.Result[Response], error) {
	traveler, err := user.FindByUUID(travelerUUID)
	if err != nil {
		return pagination.Result[Response]{}, err
	}
	rows, total, err := FindAll(params, Filters{Status: status, UserID: &traveler.ID})
	if err != nil {
		return pagination.Result[Response]{}, err
	}
	items := make([]Response, len(rows))
	for i, row := range rows {
		items[i] = ToResponse(row)
	}
	return pagination.Result[Response]{Items: items, Meta: pagination.NewMeta(total, params)}, nil
}

// Book lets a traveler book a package for themselves. It starts as pending
// until the company confirms it.
func Book(travelerUUID uuid.UUID, input TravelerInput) (Response, error) {
	traveler, err := user.FindByUUID(travelerUUID)
	if err != nil {
		return Response{}, err
	}
	pkg, err := tourpackage.FindByUUID(input.PackageUUID)
	if err != nil {
		return Response{}, fmt.Errorf("%w: package_uuid", ErrInvalidReference)
	}
	if !tourpackage.Published(pkg) {
		return Response{}, ErrUnavailable
	}
	return create(pkg, traveler, input.BookingType, input.Pax, input.Notes, StatusPending)
}

// CancelMine cancels one of the traveler's own pending or confirmed bookings.
func CancelMine(travelerUUID, bookingUUID uuid.UUID) (Response, error) {
	traveler, err := user.FindByUUID(travelerUUID)
	if err != nil {
		return Response{}, err
	}
	b, err := FindByUUID(bookingUUID)
	if err != nil {
		return Response{}, err
	}
	if b.UserID != traveler.ID {
		return Response{}, ErrNotYours
	}
	if !b.Status.HoldsSeats() {
		return Response{}, ErrFinal
	}
	if err := applyStatus(&b, StatusCancelled); err != nil {
		return Response{}, err
	}
	if err := Save(&b); err != nil {
		return Response{}, err
	}
	saved, err := FindByUUID(b.UUID)
	if err != nil {
		return Response{}, err
	}
	resp := ToResponse(saved)
	publish(ws.ActionUpdated, resp)
	notifyStatus(saved, true)
	return resp, nil
}
