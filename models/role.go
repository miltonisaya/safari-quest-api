package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string         `gorm:"uniqueIndex;not null" json:"name"`
	Code      string         `gorm:"uniqueIndex;not null" json:"code"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (r *Role) BeforeCreate(tx *gorm.DB) error {
	r.ID = uuid.New()
	return nil
}
