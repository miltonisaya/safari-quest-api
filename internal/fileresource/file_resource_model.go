package fileresource

import (
	"time"

	"safari-quest-api/internal/fileresourcetype"
	"safari-quest-api/internal/touroperator"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FileResource is one uploaded file. OperatorID is the company it belongs to.
type FileResource struct {
	ID                 uint                              `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID               uuid.UUID                         `gorm:"type:uuid;uniqueIndex;not null" json:"-"`
	Name               string                            `gorm:"not null" json:"-"`
	Code               string                            `gorm:"uniqueIndex;not null" json:"-"`
	DateUploaded       time.Time                         `gorm:"not null" json:"-"`
	FileResourceTypeID uint                              `gorm:"not null" json:"-"`
	FileResourceType   fileresourcetype.FileResourceType `gorm:"foreignKey:FileResourceTypeID" json:"-"`
	OperatorID         uint                              `gorm:"not null" json:"-"`
	Operator           touroperator.TourOperator         `gorm:"foreignKey:OperatorID" json:"-"`
	Path               string                            `gorm:"not null" json:"-"`
	CreatedAt          time.Time                         `json:"-"`
	UpdatedAt          time.Time                         `json:"-"`
	DeletedAt          gorm.DeletedAt                    `gorm:"index" json:"-"`
}

func (f *FileResource) BeforeCreate(tx *gorm.DB) error {
	if f.UUID == uuid.Nil {
		f.UUID = uuid.New()
	}
	return nil
}
