package geography

import (
	"safari-quest-api/internal/geographytype"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Geography struct {
	ID              uint                        `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID            uuid.UUID                   `gorm:"type:uuid;uniqueIndex;not null" json:"-"`
	Name            string                      `gorm:"uniqueIndex;not null" json:"-"`
	Code            string                      `gorm:"uniqueIndex;not null" json:"-"`
	GeographyTypeID uint                        `gorm:"column:geography_type_id;not null" json:"-"`
	GeographyType   geographytype.GeographyType `gorm:"foreignKey:GeographyTypeID"`
	ParentID        *uint                       `gorm:"index" json:"-"`
	Parent          *Geography                  `gorm:"foreignKey:ParentID" json:"-"`
	Children        []Geography                 `gorm:"foreignKey:ParentID" json:"-"`
	CreatedAt       time.Time                   `json:"-"`
	UpdatedAt       time.Time                   `json:"-"`
	DeletedAt       gorm.DeletedAt              `gorm:"index" json:"-"`
}

func (r *Geography) BeforeCreate(tx *gorm.DB) error {
	r.UUID = uuid.New()
	return nil
}
