package booking

import (
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
	UnitPrice      float64                   `gorm:"type:numeric(12,2);not null" json:"-"`
	TotalPrice     float64                   `gorm:"type:numeric(12,2);not null" json:"-"`
	Status         Status                    `gorm:"type:varchar(20);not null;default:pending" json:"-"`
	Notes          *string                   `gorm:"type:text" json:"-"`
	ConfirmedAt    *time.Time                `json:"-"`
	CancelledAt    *time.Time                `json:"-"`
	CreatedAt      time.Time                 `json:"-"`
	UpdatedAt      time.Time                 `json:"-"`
	DeletedAt      gorm.DeletedAt            `gorm:"index" json:"-"`
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
		return next == StatusConfirmed || next == StatusCancelled
	case StatusConfirmed:
		return next == StatusCompleted || next == StatusCancelled
	default:
		return false
	}
}

// ValidStatus reports whether s is one of the known statuses.
func ValidStatus(s Status) bool {
	switch s {
	case StatusPending, StatusConfirmed, StatusCompleted, StatusCancelled:
		return true
	}
	return false
}
