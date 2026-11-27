package agegroup

import (
	"time"

	"github.com/google/uuid"
)

// AgeGroup is an age band travellers are put in (e.g. "5–15"). The bands
// start at 0 and run without gaps; the oldest has no MaxAge.
type AgeGroup struct {
	ID        uint      `gorm:"primaryKey"`
	UUID      uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`
	Name      string    `gorm:"not null"`
	MinAge    int       `gorm:"not null"`
	MaxAge    *int
	CreatedAt time.Time
	UpdatedAt time.Time
}
