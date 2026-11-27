package booking

import (
	"time"

	"github.com/google/uuid"
)

// ID types a group leader can give.
const (
	IDPassport   = "passport"
	IDNationalID = "national_id"
)

// Traveller is one person on a booking: the lead (who booked, with ID) or a
// member. The age group's name and range are copied so later changes to the
// age groups never rewrite past bookings.
type Traveller struct {
	ID             uint      `gorm:"primaryKey"`
	UUID           uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`
	BookingID      uint      `gorm:"not null"`
	IsLead         bool      `gorm:"not null;default:false"`
	Position       int       `gorm:"not null"`
	FullName       string    `gorm:"not null"`
	Email          *string
	Phone          *string
	Nationality    string `gorm:"type:char(2);not null"`
	AgeGroupID     *uint
	AgeGroupName   string `gorm:"not null"`
	AgeMin         int    `gorm:"not null"`
	AgeMax         *int
	IDType         *string
	IDNumber       *string
	IDExpiry       *time.Time `gorm:"type:date"`
	IDDocumentPath *string    // relative to the private uploads folder; never served directly
	IDDocumentName *string
	IDDocumentType *string
	IDUploadedAt   *time.Time
	IDDeletedAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (Traveller) TableName() string { return "booking_travellers" }
