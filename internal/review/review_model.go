package review

import (
	"time"

	"safari-quest-api/internal/booking"
	"safari-quest-api/internal/tourpackage"
	"safari-quest-api/internal/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Review is a traveler's rating of a package, tied to one of their completed
// bookings (one review per booking). The package's company may reply once;
// administrators may hide abusive reviews without deleting them.
type Review struct {
	ID            uint                `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID          uuid.UUID           `gorm:"type:uuid;uniqueIndex;not null" json:"-"`
	PackageID     uint                `gorm:"not null" json:"-"`
	Package       tourpackage.Package `gorm:"foreignKey:PackageID" json:"-"`
	BookingID     uint                `gorm:"uniqueIndex;not null" json:"-"`
	Booking       booking.Booking     `gorm:"foreignKey:BookingID" json:"-"`
	UserID        uint                `gorm:"not null" json:"-"`
	User          user.User           `gorm:"foreignKey:UserID" json:"-"`
	Rating        int                 `gorm:"type:smallint;not null" json:"-"`
	Title         *string             `json:"-"`
	Comment       *string             `gorm:"type:text" json:"-"`
	OperatorReply *string             `gorm:"type:text" json:"-"`
	RepliedAt     *time.Time          `json:"-"`
	IsHidden      bool                `gorm:"not null;default:false" json:"-"`
	HiddenAt      *time.Time          `json:"-"`
	CreatedAt     time.Time           `json:"-"`
	UpdatedAt     time.Time           `json:"-"`
}

func (Review) TableName() string { return "package_reviews" }

func (r *Review) BeforeCreate(tx *gorm.DB) error {
	if r.UUID == uuid.Nil {
		r.UUID = uuid.New()
	}
	return nil
}

// ValidRating reports whether stars is a 1–5 rating.
func ValidRating(stars int) bool { return stars >= 1 && stars <= 5 }
