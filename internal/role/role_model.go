package role

import (
	"safari-quest-api/internal/authority"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	ID          uint                  `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID        uuid.UUID             `gorm:"type:uuid;uniqueIndex;not null" json:"-"`
	Name        string                `gorm:"uniqueIndex;not null" json:"-"`
	Code        string                `gorm:"uniqueIndex;not null" json:"-"`
	CreatedAt   time.Time             `json:"-"`
	UpdatedAt   time.Time             `json:"-"`
	DeletedAt   gorm.DeletedAt        `gorm:"index" json:"-"`
	Authorities []authority.Authority `gorm:"many2many:role_authorities;" json:"-"`
}

func (r *Role) BeforeCreate(tx *gorm.DB) error {
	// Saving a user replaces this association and runs the hook on the loaded role.
	// Keep the UUID the database already has.
	if r.UUID == uuid.Nil {
		r.UUID = uuid.New()
	}
	return nil
}
