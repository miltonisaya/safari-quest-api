package attraction

import (
	"safari-quest-api/database"
	"safari-quest-api/internal/touroperator"
	"safari-quest-api/pkg/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var attractionSortColumns = map[string]string{
	"name":       "name",
	"code":       "code",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

// Filters narrow the attraction list; nil fields are ignored.
type Filters struct {
	AttractionTypeUUID     *uuid.UUID
	AttractionCategoryUUID *uuid.UUID
	GeographyUUID          *uuid.UUID
	TourOperatorUUID       *uuid.UUID
}

func withRelations(db *gorm.DB) *gorm.DB {
	return db.Preload("AttractionType").Preload("AttractionCategory").Preload("Geography").Preload("TourOperators")
}

func FindAll(params pagination.Params, f Filters) ([]Attraction, int64, error) {
	var attractions []Attraction
	var total int64

	db := database.GORM_DB.Model(&Attraction{})

	if params.Search != "" {
		term := "%" + params.Search + "%"
		db = db.Where("attractions.name ILIKE ? OR attractions.code ILIKE ?", term, term)
	}
	if f.AttractionTypeUUID != nil {
		db = db.Where("attraction_type_id = (SELECT id FROM attraction_types WHERE uuid = ?)", *f.AttractionTypeUUID)
	}
	if f.AttractionCategoryUUID != nil {
		db = db.Where("attraction_category_id = (SELECT id FROM attraction_categories WHERE uuid = ?)", *f.AttractionCategoryUUID)
	}
	if f.GeographyUUID != nil {
		db = db.Where("geography_id = (SELECT id FROM geographies WHERE uuid = ?)", *f.GeographyUUID)
	}
	if f.TourOperatorUUID != nil {
		db = db.Where(`attractions.id IN (
			SELECT ato.attraction_id FROM attraction_tour_operators ato
			JOIN tour_operators t ON t.id = ato.tour_operator_id
			WHERE t.uuid = ?)`, *f.TourOperatorUUID)
	}

	db.Count(&total)

	result := withRelations(db).
		Order(params.OrderClause(attractionSortColumns, "created_at")).
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&attractions)

	return attractions, total, result.Error
}

func FindByUUID(uuid uuid.UUID) (Attraction, error) {
	var attraction Attraction
	result := withRelations(database.GORM_DB).First(&attraction, "uuid = ?", uuid)
	return attraction, result.Error
}

// Insert creates the attraction and links its tour operators in one transaction.
func Insert(attraction *Attraction, operators []touroperator.TourOperator) error {
	return database.GORM_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("AttractionType", "AttractionCategory", "Geography", "TourOperators").Create(attraction).Error; err != nil {
			return err
		}
		return tx.Session(&gorm.Session{FullSaveAssociations: false}).
			Model(attraction).Association("TourOperators").Replace(operators)
	})
}

// Save updates the attraction and replaces its tour operators in one transaction.
func Save(attraction *Attraction, operators []touroperator.TourOperator) error {
	return database.GORM_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("AttractionType", "AttractionCategory", "Geography", "TourOperators").Save(attraction).Error; err != nil {
			return err
		}
		return tx.Session(&gorm.Session{FullSaveAssociations: false}).
			Model(attraction).Association("TourOperators").Replace(operators)
	})
}

func Remove(uuid uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", uuid).Delete(&Attraction{}).Error
}
