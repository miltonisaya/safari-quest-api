package booking

import (
	"safari-quest-api/pkg/dbtypes"
	"time"

	"safari-quest-api/internal/touroperator"
	"safari-quest-api/internal/tourpackage"
	"safari-quest-api/internal/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Status is where a booking is in its life cycle.
type Status string

const (
	StatusPending   Status = "pending"
	StatusConfirmed Status = "confirmed"
	StatusCompleted Status = "completed"
	StatusCancelled Status = "cancelled"
	StatusDeclined  Status = "declined" // the company turned the request down
	StatusExpired   Status = "expired"  // not confirmed before its deadline
)

// Booking is a traveler's reservation of seats on a sale package. Trip dates
// come from the package. UnitPrice is the package's price per person at the
// time of booking, so later price changes do not alter existing bookings.
type Booking struct {
	ID             uint                      `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID           uuid.UUID                 `gorm:"type:uuid;uniqueIndex;not null" json:"-"`
	Code           string                    `gorm:"uniqueIndex;not null" json:"-"`
	PackageID      uint                      `gorm:"not null" json:"-"`
	Package        tourpackage.Package       `gorm:"foreignKey:PackageID" json:"-"`
	TourOperatorID uint                      `gorm:"not null" json:"-"` // company when booked; access follows the package's current company
	TourOperator   touroperator.TourOperator `gorm:"foreignKey:TourOperatorID" json:"-"`
	UserID         uint                      `gorm:"not null" json:"-"`
	User           user.User                 `gorm:"foreignKey:UserID" json:"-"`
	Pax            int                       `gorm:"not null" json:"-"`
	BookingType    string                    `gorm:"type:varchar(10);not null;default:shared" json:"-"` // shared or private
	UnitNumber     int                       `gorm:"not null;default:1" json:"-"`                       // which group (1…package units)
	// Pricing copied from the package when booked (see Quote).
	Pricing           string     `gorm:"type:varchar(12);not null;default:per_person" json:"-"`
	MinTravellers     int        `gorm:"not null;default:1" json:"-"`
	ChargedTravellers int        `gorm:"not null;default:1" json:"-"`
	UnitPrice         float64    `gorm:"type:numeric(12,2);not null" json:"-"`
	TotalPrice        float64    `gorm:"type:numeric(12,2);not null" json:"-"`
	Status            Status     `gorm:"type:varchar(20);not null;default:pending" json:"-"`
	Notes             *string    `gorm:"type:text" json:"-"`
	ConfirmedAt       *time.Time `json:"-"`
	CancelledAt       *time.Time `json:"-"`
	DeclineReason     *string    `gorm:"type:text" json:"-"`
	DeclinedAt        *time.Time `json:"-"`
	ExpiresAt         *time.Time `json:"-"` // a pending booking not confirmed by then expires; nil = never
	ExpiredAt         *time.Time `json:"-"`
	// Traveller details, filled in by the traveler before the company confirms.
	PreferredLanguages dbtypes.StringList `gorm:"type:jsonb;not null;default:'[]'" json:"-"`
	SpecialRequests    *string            `gorm:"type:text" json:"-"`
	DetailsSubmittedAt *time.Time         `json:"-"`
	Travellers         []Traveller        `gorm:"foreignKey:BookingID" json:"-"`
	CreatedAt          time.Time          `json:"-"`
	UpdatedAt          time.Time          `json:"-"`
	DeletedAt          gorm.DeletedAt     `gorm:"index" json:"-"`
}

func (b *Booking) BeforeCreate(tx *gorm.DB) error {
	if b.UUID == uuid.Nil {
		b.UUID = uuid.New()
	}
	return nil
}

// HoldsSeats reports whether the booking counts against the package's group size.
func (s Status) HoldsSeats() bool {
	return s == StatusPending || s == StatusConfirmed
}

// CanMoveTo reports whether a booking may change from s to next.
// Pending → confirmed or cancelled; confirmed → completed or cancelled.
// Completed and cancelled bookings are final.
func (s Status) CanMoveTo(next Status) bool {
	if s == next {
		return true
	}
	switch s {
	case StatusPending:
		return next == StatusConfirmed || next == StatusCancelled || next == StatusDeclined || next == StatusExpired
	case StatusConfirmed:
		return next == StatusCompleted || next == StatusCancelled
	default:
		return false
	}
}

// ValidStatus reports whether s is one of the known statuses.
func ValidStatus(s Status) bool {
	switch s {
	case StatusPending, StatusConfirmed, StatusCompleted, StatusCancelled, StatusDeclined, StatusExpired:
		return true
	}
	return false
}
