package tourpackage

import (
	"time"

	"safari-quest-api/internal/attraction"
	"safari-quest-api/pkg/dbtypes"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ItineraryDay is one day of a package's plan. Days are replaced as a whole
// whenever the package is saved, so they have no soft delete.
type ItineraryDay struct {
	ID            uint                   `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID          uuid.UUID              `gorm:"type:uuid;uniqueIndex;not null" json:"-"`
	PackageID     uint                   `gorm:"not null" json:"-"`
	DayNumber     int                    `gorm:"not null" json:"-"`
	Title         string                 `gorm:"not null" json:"-"`
	Description   string                 `gorm:"type:text;not null" json:"-"`
	Accommodation *string                `json:"-"`
	Meals         dbtypes.StringList     `gorm:"type:jsonb;not null;default:'[]'" json:"-"`
	AttractionID  *uint                  `json:"-"`
	Attraction    *attraction.Attraction `gorm:"foreignKey:AttractionID" json:"-"`
	CreatedAt     time.Time              `json:"-"`
	UpdatedAt     time.Time              `json:"-"`
}

func (ItineraryDay) TableName() string { return "package_itinerary_days" }

func (d *ItineraryDay) BeforeCreate(tx *gorm.DB) error {
	if d.UUID == uuid.Nil {
		d.UUID = uuid.New()
	}
	return nil
}

// Meal values accepted on an itinerary day.
var validMeals = map[string]bool{"breakfast": true, "lunch": true, "dinner": true}

// Difficulty values accepted on a package.
var validDifficulty = map[string]bool{"easy": true, "moderate": true, "challenging": true}
