package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Menu struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID        uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"uuid"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Type        MenuType  `gorm:"type:varchar(20);not null" json:"type"`
	Route       *string   `gorm:"size:255" json:"route,omitempty"`
	Icon        *string   `gorm:"size:100" json:"icon,omitempty"`
	AuthorityID uint      `gorm:"not null;index" json:"authority_id"`
	Authority   Authority `gorm:"foreignKey:AuthorityID" json:"authority"`
	SortOrder   int       `gorm:"default:0" json:"sortOrder"`
	Active      bool      `gorm:"default:true" json:"active"`

	//Recursive relationship
	ParentID  *uint     `gorm:"index" json:"parent_id,omitempty"`
	Parent    *Menu     `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children  []Menu    `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type MenuType string

const (
	MenuTypeGroup MenuType = "GROUP"
	MenuTypeLink  MenuType = "LINK"
)

func (m *Menu) BeforeCreate(tx *gorm.DB) error {
	if m.UUID == uuid.Nil {
		m.UUID = uuid.New()
	}
	return nil
}
