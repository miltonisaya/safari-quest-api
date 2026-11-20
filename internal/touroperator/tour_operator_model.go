package touroperator

import (
	"time"

	"safari-quest-api/internal/geography"
	"safari-quest-api/internal/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TourOperator is a safari company that offers trips to attractions. Its
// staff accounts are linked through the tour_operator_users pivot table.
type TourOperator struct {
	ID                 uint                 `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID               uuid.UUID            `gorm:"type:uuid;uniqueIndex;not null" json:"-"`
	Name               string               `gorm:"uniqueIndex;not null" json:"-"` // trading name shown to travelers
	Code               string               `gorm:"uniqueIndex;not null" json:"-"`
	LegalName          *string              `json:"-"` // registered company name
	RegistrationNumber *string              `gorm:"uniqueIndex" json:"-"`
	LicenseNumber      *string              `gorm:"uniqueIndex" json:"-"` // tourism (e.g. TALA) licence
	Email              string               `gorm:"not null" json:"-"`
	Phone              string               `gorm:"not null" json:"-"`
	Website            *string              `json:"-"`
	Address            *string              `gorm:"type:text" json:"-"`
	GeographyID        *uint                `json:"-"` // where the operator is based
	Geography          *geography.Geography `gorm:"foreignKey:GeographyID" json:"-"`
	Description        *string              `gorm:"type:text" json:"-"`
	IsVerified         bool                 `gorm:"default:false;not null" json:"-"`
	IsActive           bool                 `gorm:"default:true;not null" json:"-"`
	CreatedAt          time.Time            `json:"-"`
	UpdatedAt          time.Time            `json:"-"`
	DeletedAt          gorm.DeletedAt       `gorm:"index" json:"-"`
	Users              []user.User          `gorm:"many2many:tour_operator_users;" json:"-"`
}

func (t *TourOperator) BeforeCreate(tx *gorm.DB) error {
	t.UUID = uuid.New()
	return nil
}
