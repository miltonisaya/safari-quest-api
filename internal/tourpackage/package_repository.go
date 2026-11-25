package tourpackage

import (
	"time"

	"safari-quest-api/database"
	"safari-quest-api/internal/attraction"
	"safari-quest-api/internal/fileresource"
	"safari-quest-api/pkg/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var sortColumns = map[string]string{
	"name":         "packages.name",
	"start_date":   "packages.start_date",
	"end_date":     "packages.end_date",
	"pax":          "packages.pax",
	"price":        "packages.price",
	"package_type": "packages.package_type",
	"created_at":   "packages.created_at",
	"updated_at":   "packages.updated_at",
}

func withRelations(db *gorm.DB) *gorm.DB {
	return db.Preload("Operator").
		Preload("FileResources.FileResourceType").Preload("FileResources.Operator").
		Preload("Attractions").
		Preload("Itinerary", func(db *gorm.DB) *gorm.DB { return db.Order("day_number ASC") }).
		Preload("Itinerary.Attraction")
}

func FindAll(params pagination.Params, operatorIDs []uint, unrestricted bool) ([]Package, int64, error) {
	var items []Package
	var total int64

	db := database.GORM_DB.Model(&Package{})
	if !unrestricted {
		if len(operatorIDs) == 0 {
			db = db.Where("1 = 0")
		} else {
			db = db.Where("packages.operator_id IN ?", operatorIDs)
		}
	}
	if params.Search != "" {
		term := "%" + params.Search + "%"
		db = db.Where(
			"packages.name ILIKE ? OR packages.description ILIKE ? OR packages.package_type ILIKE ?",
			term, term, term,
		)
	}
	db.Count(&total)

	result := withRelations(db).
		Order(params.OrderClause(sortColumns, "packages.created_at")).
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&items)
	return items, total, result.Error
}

func FindByUUID(id uuid.UUID) (Package, error) {
	var item Package
	result := withRelations(database.GORM_DB).First(&item, "uuid = ?", id)
	return item, result.Error
}

// Links are what a package points at besides its own columns.
type Links struct {
	Files       []fileresource.FileResource
	Attractions []attraction.Attraction
	Itinerary   []ItineraryDay
}

var omitRelations = []string{"Operator", "FileResources", "Attractions", "Itinerary"}

// writeLinks replaces the package's files, attractions and itinerary days.
func writeLinks(tx *gorm.DB, item *Package, links Links) error {
	assoc := tx.Session(&gorm.Session{FullSaveAssociations: false}).Model(item)
	if err := assoc.Association("FileResources").Replace(links.Files); err != nil {
		return err
	}
	if err := tx.Session(&gorm.Session{FullSaveAssociations: false}).Model(item).
		Association("Attractions").Replace(links.Attractions); err != nil {
		return err
	}
	if err := tx.Where("package_id = ?", item.ID).Delete(&ItineraryDay{}).Error; err != nil {
		return err
	}
	for i := range links.Itinerary {
		day := links.Itinerary[i]
		day.ID, day.UUID, day.PackageID, day.Attraction = 0, uuid.Nil, item.ID, nil
		if err := tx.Omit("Attraction").Create(&day).Error; err != nil {
			return err
		}
	}
	return nil
}

func Insert(item *Package, links Links) error {
	return database.GORM_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit(omitRelations...).Create(item).Error; err != nil {
			return err
		}
		return writeLinks(tx, item, links)
	})
}

func Save(item *Package, links Links) error {
	return database.GORM_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit(omitRelations...).Save(item).Error; err != nil {
			return err
		}
		return writeLinks(tx, item, links)
	})
}

// Rating is the average stars and count of a package's visible reviews.
type Rating struct {
	PackageID uint
	Average   float64
	Count     int
}

// RatingsFor summarises visible reviews for the given packages.
func RatingsFor(packageIDs []uint) (map[uint]Rating, error) {
	out := make(map[uint]Rating, len(packageIDs))
	if len(packageIDs) == 0 {
		return out, nil
	}
	var rows []Rating
	err := database.GORM_DB.Table("package_reviews").
		Select("package_id, ROUND(AVG(rating)::numeric, 1)::float8 AS average, COUNT(*) AS count").
		Where("package_id IN ? AND is_hidden = FALSE", packageIDs).
		Group("package_id").
		Scan(&rows).Error
	for _, r := range rows {
		out[r.PackageID] = r
	}
	return out, err
}

func Remove(id uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", id).Delete(&Package{}).Error
}

// PublishedOperator limits packages to companies that are active and
// verified. Self-registered companies start unverified, so their packages stay
// off the public site and out of traveler booking until an administrator
// verifies them.
const PublishedOperator = "packages.operator_id IN (SELECT id FROM tour_operators WHERE is_active = TRUE AND is_verified = TRUE AND deleted_at IS NULL)"

// Published reports whether a package's company is active and verified.
func Published(p Package) bool {
	return p.Operator.IsActive && p.Operator.IsVerified
}

// CatalogFilters narrow the public catalog; zero values are ignored.
type CatalogFilters struct {
	Category     string  // matches package_type, case-insensitive
	EcoCertified bool    // only eco-certified packages
	MaxPrice     float64 // price per person ceiling
	Search       string  // name, description, location or package type
	Sort         string  // soonest (default), price-asc, price-desc, newest
}

// FindCatalog lists packages that have not started yet, for the public site.
func FindCatalog(f CatalogFilters, limit, offset int) ([]Package, int64, error) {
	db := database.GORM_DB.Model(&Package{}).
		Where("packages.start_date >= ?", time.Now().Format("2006-01-02")).
		Where(PublishedOperator)
	if f.Category != "" {
		db = db.Where("LOWER(packages.package_type) = LOWER(?)", f.Category)
	}
	if f.EcoCertified {
		db = db.Where("packages.eco_certified = TRUE")
	}
	if f.MaxPrice > 0 {
		db = db.Where("packages.price <= ?", f.MaxPrice)
	}
	if f.Search != "" {
		term := "%" + f.Search + "%"
		db = db.Where("packages.name ILIKE ? OR packages.description ILIKE ? OR packages.location ILIKE ? OR packages.package_type ILIKE ?",
			term, term, term, term)
	}
	var total int64
	db.Count(&total)

	order := "packages.start_date ASC"
	switch f.Sort {
	case "price-asc":
		order = "packages.price ASC"
	case "price-desc":
		order = "packages.price DESC"
	case "newest":
		order = "packages.created_at DESC"
	}
	var items []Package
	err := withRelations(db).Order(order).Limit(limit).Offset(offset).Find(&items).Error
	return items, total, err
}
