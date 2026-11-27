package booking

import (
	"time"

	"github.com/google/uuid"
)

// Message is one chat message on a booking's conversation between the
// traveler (FromCompany false) and the company's staff (FromCompany true).
type Message struct {
	ID          uint      `gorm:"primaryKey"`
	UUID        uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`
	BookingID   uint      `gorm:"not null"`
	SenderID    *uint
	FromCompany bool   `gorm:"not null"`
	Body        string `gorm:"not null"`
	ReadAt      *time.Time
	CreatedAt   time.Time
}

func (Message) TableName() string { return "booking_messages" }
