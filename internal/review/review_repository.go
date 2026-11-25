package review

import (
	"safari-quest-api/database"
	"safari-quest-api/pkg/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var sortColumns = map[string]string{
	"rating":     "package_reviews.rating",
	"created_at": "package_reviews.created_at",
	"updated_at": "package_reviews.updated_at",
}

// Filters narrow a review list; zero values are ignored.
type Filters struct {
	PackageID      *uint
	UserID         *uint
	OnlyVisible    bool
	Admin          bool   // with OperatorIDs: company scope for staff lists
	OperatorIDs    []uint // limits to packages of these companies unless Admin
	ScopeToCompany bool
}

func withRelations(db *gorm.DB) *gorm.DB {
	return db.Preload("Package").Preload("Package.Operator").Preload("Booking").Preload("User")
}

func FindAll(params pagination.Params, f Filters) ([]Review, int64, error) {
	var items []Review
	var total int64

	db := database.GORM_DB.Model(&Review{})
	if f.PackageID != nil {
		db = db.Where("package_reviews.package_id = ?", *f.PackageID)
	}
	if f.UserID != nil {
		db = db.Where("package_reviews.user_id = ?", *f.UserID)
	}
	if f.OnlyVisible {
		db = db.Where("package_reviews.is_hidden = FALSE")
	}
	if f.ScopeToCompany && !f.Admin {
		if len(f.OperatorIDs) == 0 {
			db = db.Where("1 = 0")
		} else {
			db = db.Where("package_reviews.package_id IN (SELECT id FROM packages WHERE operator_id IN ?)", f.OperatorIDs)
		}
	}
	if params.Search != "" {
		term := "%" + params.Search + "%"
		db = db.Where("package_reviews.title ILIKE ? OR package_reviews.comment ILIKE ?", term, term)
	}
	db.Count(&total)

	result := withRelations(db).
		Order(params.OrderClause(sortColumns, "package_reviews.created_at")).
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&items)
	return items, total, result.Error
}

func FindByUUID(id uuid.UUID) (Review, error) {
	var item Review
	result := withRelations(database.GORM_DB).First(&item, "uuid = ?", id)
	return item, result.Error
}

// ExistsForBooking reports whether the booking already has a review.
func ExistsForBooking(bookingID uint) (bool, error) {
	var count int64
	err := database.GORM_DB.Model(&Review{}).Where("booking_id = ?", bookingID).Count(&count).Error
	return count > 0, err
}

var omitRelations = []string{"Package", "Booking", "User"}

func Insert(r *Review) error { return database.GORM_DB.Omit(omitRelations...).Create(r).Error }

func Save(r *Review) error { return database.GORM_DB.Omit(omitRelations...).Save(r).Error }

func Remove(id uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", id).Delete(&Review{}).Error
}
