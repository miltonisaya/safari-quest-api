package tourpackage

import (
	"time"

	"safari-quest-api/internal/fileresource"
	"safari-quest-api/internal/touroperator"

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
