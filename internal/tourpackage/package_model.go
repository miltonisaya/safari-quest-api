package tourpackage

import (
	"time"

	"safari-quest-api/internal/attraction"
	"safari-quest-api/internal/fileresource"
	"safari-quest-api/internal/touroperator"
	"safari-quest-api/pkg/dbtypes"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Package is a trip a tour company sells to travellers.
type Package struct {
	ID            uint                        `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID          uuid.UUID                   `gorm:"type:uuid;uniqueIndex;not null" json:"-"`
	OperatorID    uint                        `gorm:"not null" json:"-"`
	Operator      touroperator.TourOperator   `gorm:"foreignKey:OperatorID" json:"-"`
	Name          string                      `gorm:"not null" json:"-"`
	Description   string                      `gorm:"type:text;not null" json:"-"`
	StartDate     time.Time                   `gorm:"type:date;not null" json:"-"`
	EndDate       time.Time                   `gorm:"type:date;not null" json:"-"`
	Pax           int                         `gorm:"not null" json:"-"`
	Price         float64                     `gorm:"type:numeric(12,2);not null" json:"-"`
	PackageType   string                      `gorm:"not null" json:"-"`
	Highlights    dbtypes.StringList          `gorm:"type:jsonb;not null;default:'[]'" json:"-"`
	Inclusions    dbtypes.StringList          `gorm:"type:jsonb;not null;default:'[]'" json:"-"`
	Exclusions    dbtypes.StringList          `gorm:"type:jsonb;not null;default:'[]'" json:"-"`
	Difficulty    *string                     `gorm:"type:varchar(20)" json:"-"` // easy, moderate or challenging
	MeetingPoint  *string                     `json:"-"`
	Location      *string                     `json:"-"` // e.g. "Serengeti, Tanzania"
	Currency      string                      `gorm:"type:char(3);not null;default:USD" json:"-"`
	EcoCertified  bool                        `gorm:"not null;default:false" json:"-"`
	Attractions   []attraction.Attraction     `gorm:"many2many:package_attractions;" json:"-"`
	Itinerary     []ItineraryDay              `gorm:"foreignKey:PackageID" json:"-"`
	FileResources []fileresource.FileResource `gorm:"many2many:package_file_resources;" json:"-"`
	CreatedAt     time.Time                   `json:"-"`
	UpdatedAt     time.Time                   `json:"-"`
	DeletedAt     gorm.DeletedAt              `gorm:"index" json:"-"`
}

func (p *Package) BeforeCreate(tx *gorm.DB) error {
	if p.UUID == uuid.Nil {
		p.UUID = uuid.New()
	}
	return nil
}
