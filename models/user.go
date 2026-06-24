package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID              uint           `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID            uuid.UUID      `gorm:"type:uuid;uniqueIndex;not null" json:"-"`
	FirstName       string         `gorm:"not null" json:"-"`
	MiddleName      string         `gorm:"type:text" json:"-"`
	LastName        string         `gorm:"not null" json:"-"`
	Email           string         `gorm:"uniqueIndex;not null" json:"-"`
	Password        string         `gorm:"not null" json:"-"`
	Sex             string         `gorm:"not null" json:"-"`
	Mobile          string         `gorm:"not null" json:"-"`
	Address         string         `gorm:"not null" json:"-"`
	IsActive        bool           `gorm:"default:true;not null" json:"-"`
	EmailVerifiedAt *time.Time     `json:"-"`
	CreatedAt       time.Time      `json:"-"`
	UpdatedAt       time.Time      `json:"-"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	Roles           []Role         `gorm:"many2many:user_roles;" json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.UUID = uuid.New()
	return nil
}
