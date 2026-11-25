package review

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"safari-quest-api/internal/booking"
	"safari-quest-api/internal/companyscope"
	"safari-quest-api/internal/tourpackage"
	"safari-quest-api/internal/user"
	"safari-quest-api/pkg/pagination"
	ws "safari-quest-api/websocket"

	"github.com/google/uuid"
)

var (
	ErrInvalidReference = errors.New("invalid reference")
	ErrNotYours         = errors.New("You can only review and edit your own trips")
	ErrNotCompleted     = errors.New("You can review a trip once it is completed")
	ErrAlreadyReviewed  = errors.New("You have already reviewed this trip")
	ErrInvalidRating    = errors.New("Rating must be between 1 and 5 stars")
	ErrOutsideCompany   = errors.New("You can only reply to reviews of your company's packages")
)

// Input is a traveler's review of one of their completed bookings.
type Input struct {
	BookingUUID uuid.UUID `json:"booking_uuid" binding:"required"`
	Rating      int       `json:"rating" binding:"required,min=1,max=5"`
	Title       *string   `json:"title" binding:"omitempty,max=150"`
	Comment     *string   `json:"comment" binding:"omitempty,max=4000"`
}

// UpdateInput changes the traveler's own review.
type UpdateInput struct {
	Rating  int     `json:"rating" binding:"required,min=1,max=5"`
	Title   *string `json:"title" binding:"omitempty,max=150"`
	Comment *string `json:"comment" binding:"omitempty,max=4000"`
}

// ReplyInput is the company's public answer; empty removes it.
type ReplyInput struct {
	Reply *string `json:"reply" binding:"omitempty,max=4000"`
}

// ModerationInput hides or shows a review.
type ModerationInput struct {
	Hidden bool `json:"hidden"`
}

type PackageRef struct {
	UUID uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
}

type Response struct {
	UUID          uuid.UUID  `json:"uuid"`
	Package       PackageRef `json:"package"`
	BookingUUID   uuid.UUID  `json:"booking_uuid"`
	Reviewer      string     `json:"reviewer"` // first name and last initial
	TripDate      string     `json:"trip_date"`
	Rating        int        `json:"rating"`
	Title         *string    `json:"title"`
	Comment       *string    `json:"comment"`
	OperatorReply *string    `json:"operator_reply"`
	RepliedAt     *time.Time `json:"replied_at"`
	IsHidden      bool       `json:"is_hidden"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type Summary struct {
	Average float64 `json:"average"`
	Count   int     `json:"count"`
}

// PackageList is a package's visible reviews with its rating summary.
type PackageList struct {
	Items   []Response      `json:"items"`
	Meta    pagination.Meta `json:"meta"`
	Summary Summary         `json:"summary"`
}

// reviewerName shows "Amani T." so public reviews do not expose full names.
func reviewerName(u user.User) string {
	name := strings.TrimSpace(u.FirstName)
	if last := strings.TrimSpace(u.LastName); last != "" {
		name += " " + strings.ToUpper(last[:1]) + "."
	}
	if name == "" {
		return "Traveler"
	}
	return name
}

func ToResponse(r Review) Response {
	return Response{
		UUID:          r.UUID,
		Package:       PackageRef{UUID: r.Package.UUID, Name: r.Package.Name},
		BookingUUID:   r.Booking.UUID,
		Reviewer:      reviewerName(r.User),
		TripDate:      r.Package.StartDate.Format("2006-01-02"),
		Rating:        r.Rating,
		Title:         r.Title,
		Comment:       r.Comment,
		OperatorReply: r.OperatorReply,
		RepliedAt:     r.RepliedAt,
		IsHidden:      r.IsHidden,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

func clean(s *string) *string {
	if s == nil {
		return nil
	}
	if t := strings.TrimSpace(*s); t != "" {
		return &t
	}
	return nil
}

func toList(rows []Review, total int64, params pagination.Params) pagination.Result[Response] {
	items := make([]Response, len(rows))
	for i, r := range rows {
		items[i] = ToResponse(r)
	}
	return pagination.Result[Response]{Items: items, Meta: pagination.NewMeta(total, params)}
}

func reload(id uuid.UUID, action ws.Action) (Response, error) {
	saved, err := FindByUUID(id)
	if err != nil {
		return Response{}, err
	}
	resp := ToResponse(saved)
	ws.GlobalHub.Publish(ws.Event{Entity: "review", Action: action, Payload: resp})
	return resp, nil
}

// ── Anyone signed in ────────────────────────────────────────────────────────

// ForPackage lists a package's visible reviews, newest first by default.
func ForPackage(packageUUID uuid.UUID, params pagination.Params) (PackageList, error) {
	pkg, err := tourpackage.FindByUUID(packageUUID)
	if err != nil {
		return PackageList{}, fmt.Errorf("%w: package_uuid", ErrInvalidReference)
	}
	if params.SortBy == "" {
		params.SortDir = "desc"
	}
	rows, total, err := FindAll(params, Filters{PackageID: &pkg.ID, OnlyVisible: true})
	if err != nil {
		return PackageList{}, err
	}
	ratings, err := tourpackage.RatingsFor([]uint{pkg.ID})
	if err != nil {
		return PackageList{}, err
	}
	list := toList(rows, total, params)
	r := ratings[pkg.ID]
	return PackageList{Items: list.Items, Meta: list.Meta, Summary: Summary{Average: r.Average, Count: r.Count}}, nil
}

// ── Travelers ───────────────────────────────────────────────────────────────

func Mine(travelerUUID uuid.UUID, params pagination.Params) (pagination.Result[Response], error) {
	traveler, err := user.FindByUUID(travelerUUID)
	if err != nil {
		return pagination.Result[Response]{}, err
	}
	rows, total, err := FindAll(params, Filters{UserID: &traveler.ID})
	if err != nil {
		return pagination.Result[Response]{}, err
	}
	return toList(rows, total, params), nil
}

// Create reviews one of the traveler's own completed bookings, once.
func Create(travelerUUID uuid.UUID, input Input) (Response, error) {
	if !ValidRating(input.Rating) {
		return Response{}, ErrInvalidRating
	}
	traveler, err := user.FindByUUID(travelerUUID)
	if err != nil {
		return Response{}, err
	}
	b, err := booking.FindByUUID(input.BookingUUID)
	if err != nil {
		return Response{}, fmt.Errorf("%w: booking_uuid", ErrInvalidReference)
	}
	if b.UserID != traveler.ID {
		return Response{}, ErrNotYours
	}
	if b.Status != booking.StatusCompleted {
		return Response{}, ErrNotCompleted
	}
	exists, err := ExistsForBooking(b.ID)
	if err != nil {
		return Response{}, err
	}
	if exists {
		return Response{}, ErrAlreadyReviewed
	}
	r := Review{
		PackageID: b.PackageID,
		BookingID: b.ID,
		UserID:    traveler.ID,
		Rating:    input.Rating,
		Title:     clean(input.Title),
		Comment:   clean(input.Comment),
	}
	if err := Insert(&r); err != nil {
		return Response{}, err
	}
	return reload(r.UUID, ws.ActionCreated)
}

func ownReview(travelerUUID, reviewUUID uuid.UUID) (Review, error) {
	traveler, err := user.FindByUUID(travelerUUID)
	if err != nil {
		return Review{}, err
	}
	r, err := FindByUUID(reviewUUID)
	if err != nil {
		return Review{}, err
	}
	if r.UserID != traveler.ID {
		return Review{}, ErrNotYours
	}
	return r, nil
}

// UpdateMine edits the traveler's own review. A company reply is kept.
func UpdateMine(travelerUUID, reviewUUID uuid.UUID, input UpdateInput) (Response, error) {
	if !ValidRating(input.Rating) {
		return Response{}, ErrInvalidRating
	}
	r, err := ownReview(travelerUUID, reviewUUID)
	if err != nil {
		return Response{}, err
	}
	r.Rating = input.Rating
	r.Title = clean(input.Title)
	r.Comment = clean(input.Comment)
	if err := Save(&r); err != nil {
		return Response{}, err
	}
	return reload(r.UUID, ws.ActionUpdated)
}

func DeleteMine(travelerUUID, reviewUUID uuid.UUID) error {
	r, err := ownReview(travelerUUID, reviewUUID)
	if err != nil {
		return err
	}
	if err := Remove(r.UUID); err != nil {
		return err
	}
	ws.GlobalHub.Publish(ws.Event{Entity: "review", Action: ws.ActionDeleted, Payload: map[string]string{"uuid": r.UUID.String()}})
	return nil
}

// ── Company staff and administrators ────────────────────────────────────────

// ForCompanies lists reviews of the caller's packages, hidden ones included.
func ForCompanies(scope companyscope.Scope, packageUUID *uuid.UUID, params pagination.Params) (pagination.Result[Response], error) {
	f := Filters{ScopeToCompany: true, Admin: scope.Admin, OperatorIDs: scope.OperatorIDs}
	if packageUUID != nil {
		pkg, err := tourpackage.FindByUUID(*packageUUID)
		if err != nil {
			return pagination.Result[Response]{}, fmt.Errorf("%w: package_uuid", ErrInvalidReference)
		}
		f.PackageID = &pkg.ID
	}
	rows, total, err := FindAll(params, f)
	if err != nil {
		return pagination.Result[Response]{}, err
	}
	return toList(rows, total, params), nil
}

// Reply sets or clears the company's public answer to a review.
func Reply(reviewUUID uuid.UUID, input ReplyInput, scope companyscope.Scope) (Response, error) {
	r, err := FindByUUID(reviewUUID)
	if err != nil {
		return Response{}, err
	}
	if !scope.Allows(r.Package.OperatorID) {
		return Response{}, ErrOutsideCompany
	}
	r.OperatorReply = clean(input.Reply)
	if r.OperatorReply == nil {
		r.RepliedAt = nil
	} else {
		now := time.Now()
		r.RepliedAt = &now
	}
	if err := Save(&r); err != nil {
		return Response{}, err
	}
	return reload(r.UUID, ws.ActionUpdated)
}

// Moderate hides or shows a review. Only granted to administrators.
func Moderate(reviewUUID uuid.UUID, input ModerationInput) (Response, error) {
	r, err := FindByUUID(reviewUUID)
	if err != nil {
		return Response{}, err
	}
	r.IsHidden = input.Hidden
	if input.Hidden {
		now := time.Now()
		r.HiddenAt = &now
	} else {
		r.HiddenAt = nil
	}
	if err := Save(&r); err != nil {
		return Response{}, err
	}
	return reload(r.UUID, ws.ActionUpdated)
}
