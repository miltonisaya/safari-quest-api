package authority

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Authority struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID      uuid.UUID      `gorm:"type:uuid;uniqueIndex;not null" json:"uuid"`
	Name      string         `gorm:"uniqueIndex;not null" json:"name"`
	Code      string         `gorm:"uniqueIndex;not null" json:"code"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (a *Authority) BeforeCreate(tx *gorm.DB) error {
	if a.UUID == uuid.Nil {
		a.UUID = uuid.New()
	}
	return nil
}
