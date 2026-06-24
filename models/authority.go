package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Authority struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID      uuid.UUID      `gorm:"type:uuid;uniqueIndex;not null" json:"-"`
	Name      string         `gorm:"uniqueIndex;not null" json:"-"`
	Code      string         `gorm:"uniqueIndex;not null" json:"-"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Roles     []Role         `gorm:"many2many:role_authorities;" json:"-"`
}

func (a *Authority) BeforeCreate(tx *gorm.DB) error {
	a.UUID = uuid.New()
	return nil
}
