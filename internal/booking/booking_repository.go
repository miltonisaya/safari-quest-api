package booking

import (
	"errors"
	"time"

	"safari-quest-api/database"
	"safari-quest-api/internal/tourpackage"
	"safari-quest-api/internal/user"
	"safari-quest-api/pkg/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrNotEnoughSeats is returned when the package cannot take the requested guests.
var ErrNotEnoughSeats = errors.New("Not enough seats left on this package")

// TravelerRoleCode identifies accounts that book trips for themselves.
const TravelerRoleCode = "TRAVELER"

var sortColumns = map[string]string{
	"code":        "bookings.code",
	"pax":         "bookings.pax",
	"total_price": "bookings.total_price",
	"status":      "bookings.status",
	"created_at":  "bookings.created_at",
	"updated_at":  "bookings.updated_at",
}

// Filters narrow a booking list; empty fields are ignored.
type Filters struct {
	Status      Status
	PackageUUID *uuid.UUID
	// Scope: Admin sees everything, otherwise OperatorIDs limits the companies.
	Admin       bool
	OperatorIDs []uint
	// UserID, when set, limits the list to one traveler's bookings.
	UserID *uint
}

func withRelations(db *gorm.DB) *gorm.DB {
	return db.Preload("Package").Preload("TourOperator").Preload("User")
}

func FindAll(params pagination.Params, f Filters) ([]Booking, int64, error) {
	var items []Booking
	var total int64

	db := database.GORM_DB.Model(&Booking{})
	if f.UserID != nil {
		db = db.Where("bookings.user_id = ?", *f.UserID)
	} else if !f.Admin {
		if len(f.OperatorIDs) == 0 {
			db = db.Where("1 = 0")
		} else {
			// By the package's current company, so bookings follow a package moved to another company.
			db = db.Where("bookings.package_id IN (SELECT id FROM packages WHERE operator_id IN ?)", f.OperatorIDs)
		}
	}
	if f.Status != "" {
		db = db.Where("bookings.status = ?", f.Status)
	}
	if f.PackageUUID != nil {
		db = db.Where("bookings.package_id = (SELECT id FROM packages WHERE uuid = ?)", *f.PackageUUID)
	}
	if params.Search != "" {
		term := "%" + params.Search + "%"
		db = db.Joins("JOIN users ON users.id = bookings.user_id").
			Joins("JOIN packages ON packages.id = bookings.package_id").
			Where("bookings.code ILIKE ? OR packages.name ILIKE ? OR users.first_name ILIKE ? OR users.last_name ILIKE ? OR users.email ILIKE ?",
				term, term, term, term, term)
	}
	db.Count(&total)

	result := withRelations(db).
		Order(params.OrderClause(sortColumns, "bookings.created_at")).
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&items)
	return items, total, result.Error
}

func FindByUUID(id uuid.UUID) (Booking, error) {
	var item Booking
	result := withRelations(database.GORM_DB).First(&item, "uuid = ?", id)
	return item, result.Error
}

func CodeExists(code string) (bool, error) {
	var count int64
	err := database.GORM_DB.Unscoped().Model(&Booking{}).Where("code = ?", code).Count(&count).Error
	return count > 0, err
}

// seatsHeld sums the guests of bookings that still hold seats on the package,
// leaving out excludeID (the booking being edited).
func seatsHeld(tx *gorm.DB, packageID, excludeID uint) (int, error) {
	var held int
	err := tx.Model(&Booking{}).
		Where("package_id = ? AND id <> ? AND status IN ?", packageID, excludeID, []Status{StatusPending, StatusConfirmed}).
		Select("COALESCE(SUM(pax), 0)").Scan(&held).Error
	return held, err
}

// saveChecked writes the booking inside a transaction that locks the package
// row, so concurrent bookings cannot oversell it. Seats are checked only when
// the booking holds seats after the change.
func saveChecked(b *Booking, create bool) error {
	return database.GORM_DB.Transaction(func(tx *gorm.DB) error {
		var pkg tourpackage.Package
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&pkg, b.PackageID).Error; err != nil {
			return err
		}
		if b.Status.HoldsSeats() {
			held, err := seatsHeld(tx, pkg.ID, b.ID)
			if err != nil {
				return err
			}
			if held+b.Pax > pkg.Pax {
				return ErrNotEnoughSeats
			}
		}
		omit := []string{"Package", "TourOperator", "User"}
		if create {
			return tx.Omit(omit...).Create(b).Error
		}
		return tx.Omit(omit...).Save(b).Error
	})
}

func Insert(b *Booking) error { return saveChecked(b, true) }

func Save(b *Booking) error { return saveChecked(b, false) }

func Remove(id uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", id).Delete(&Booking{}).Error
}

// BookablePackage is a package that has not started and still has seats.
type BookablePackage struct {
	tourpackage.Package
	SeatsLeft int
}

// FindBookable lists packages starting today or later with seats left.
func FindBookable(params pagination.Params) ([]BookablePackage, int64, error) {
	held := database.GORM_DB.Model(&Booking{}).
		Select("package_id, SUM(pax) AS pax").
		Where("status IN ?", []Status{StatusPending, StatusConfirmed}).
		Group("package_id")

	db := database.GORM_DB.Model(&tourpackage.Package{}).
		Joins("LEFT JOIN (?) AS held ON held.package_id = packages.id", held).
		Where("packages.start_date >= ?", time.Now().Format("2006-01-02")).
		Where(tourpackage.PublishedOperator).
		Where("packages.pax - COALESCE(held.pax, 0) > 0")
	if params.Search != "" {
		term := "%" + params.Search + "%"
		db = db.Where("packages.name ILIKE ? OR packages.description ILIKE ? OR packages.package_type ILIKE ?", term, term, term)
	}

	var total int64
	db.Count(&total)

	var packages []tourpackage.Package
	if err := db.Preload("Operator").
		Order("packages.start_date ASC").
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&packages).Error; err != nil {
		return nil, 0, err
	}

	items := make([]BookablePackage, len(packages))
	for i, p := range packages {
		taken, err := seatsHeld(database.GORM_DB, p.ID, 0)
		if err != nil {
			return nil, 0, err
		}
		items[i] = BookablePackage{Package: p, SeatsLeft: p.Pax - taken}
	}
	return items, total, nil
}

// FindTravelers searches active accounts holding the Traveler role.
func FindTravelers(search string, limit int) ([]user.User, error) {
	db := database.GORM_DB.Model(&user.User{}).
		Joins("JOIN user_roles ur ON ur.user_id = users.id").
		Joins("JOIN roles r ON r.id = ur.role_id AND r.code = ?", TravelerRoleCode).
		Where("users.is_active = ?", true)
	if search != "" {
		term := "%" + search + "%"
		db = db.Where("users.first_name ILIKE ? OR users.last_name ILIKE ? OR users.email ILIKE ?", term, term, term)
	}
	var travelers []user.User
	err := db.Order("users.first_name, users.last_name").Limit(limit).Find(&travelers).Error
	return travelers, err
}

// SeatsHeldFor sums pending and confirmed guests per package.
func SeatsHeldFor(packageIDs []uint) (map[uint]int, error) {
	out := make(map[uint]int, len(packageIDs))
	if len(packageIDs) == 0 {
		return out, nil
	}
	var rows []struct {
		PackageID uint
		Pax       int
	}
	err := database.GORM_DB.Model(&Booking{}).
		Select("package_id, COALESCE(SUM(pax), 0) AS pax").
		Where("package_id IN ? AND status IN ?", packageIDs, []Status{StatusPending, StatusConfirmed}).
		Group("package_id").Scan(&rows).Error
	for _, r := range rows {
		out[r.PackageID] = r.Pax
	}
	return out, err
}
