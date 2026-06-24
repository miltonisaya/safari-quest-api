package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID        uuid.UUID      `gorm:"type:uuid;uniqueIndex;not null" json:"-"`
	Name        string         `gorm:"uniqueIndex;not null" json:"-"`
	Code        string         `gorm:"uniqueIndex;not null" json:"-"`
	CreatedAt   time.Time      `json:"-"`
	UpdatedAt   time.Time      `json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Authorities []Authority    `gorm:"many2many:role_authorities;" json:"-"`
	Users       []User         `gorm:"many2many:user_roles;" json:"-"`
}

func (r *Role) BeforeCreate(tx *gorm.DB) error {
	r.UUID = uuid.New()
	return nil
}
