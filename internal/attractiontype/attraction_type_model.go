package attractionType

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AttractionType struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID      uuid.UUID      `gorm:"type:uuid;uniqueIndex;not null" json:"-"`
	Name      string         `gorm:"uniqueIndex;not null" json:"name"`
	Code      string         `gorm:"uniqueIndex;not null" json:"code"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (r *AttractionType) BeforeCreate(tx *gorm.DB) error {
	r.UUID = uuid.New()
	return nil
}
