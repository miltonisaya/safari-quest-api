package attraction

import (
	"time"

	"safari-quest-api/internal/attractioncategory"
	attractiontype "safari-quest-api/internal/attractiontype"
	"safari-quest-api/internal/geography"
	"safari-quest-api/internal/touroperator"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Attraction is a place travelers visit, such as a national park or a beach.
// Many tour operators can offer trips to the same attraction.
type Attraction struct {
	ID                   uint                                  `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID                 uuid.UUID                             `gorm:"type:uuid;uniqueIndex;not null" json:"-"`
	Name                 string                                `gorm:"uniqueIndex;not null" json:"-"`
	Code                 string                                `gorm:"uniqueIndex;not null" json:"-"`
	Description          *string                               `gorm:"type:text" json:"-"`
	AttractionTypeID     uint                                  `gorm:"not null" json:"-"`
	AttractionType       attractiontype.AttractionType         `gorm:"foreignKey:AttractionTypeID" json:"-"`
	AttractionCategoryID uint                                  `gorm:"not null" json:"-"`
	AttractionCategory   attractioncategory.AttractionCategory `gorm:"foreignKey:AttractionCategoryID" json:"-"`
	GeographyID          uint                                  `gorm:"not null" json:"-"`
	Geography            geography.Geography                   `gorm:"foreignKey:GeographyID" json:"-"`
	Latitude             *float64                              `gorm:"type:numeric(9,6)" json:"-"`
	Longitude            *float64                              `gorm:"type:numeric(9,6)" json:"-"`
	IsActive             bool                                  `gorm:"default:true;not null" json:"-"`
	CreatedAt            time.Time                             `json:"-"`
	UpdatedAt            time.Time                             `json:"-"`
	DeletedAt            gorm.DeletedAt                        `gorm:"index" json:"-"`
	TourOperators        []touroperator.TourOperator           `gorm:"many2many:attraction_tour_operators;" json:"-"`
}

func (a *Attraction) BeforeCreate(tx *gorm.DB) error {
	a.UUID = uuid.New()
	return nil
}
