package attractioncategory

import (
	"safari-quest-api/database"
	"safari-quest-api/pkg/pagination"

	"github.com/google/uuid"
)

var attractionCategorySortColumns = map[string]string{
	"name":       "name",
	"code":       "code",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

func FindAll(params pagination.Params) ([]AttractionCategory, int64, error) {
	var categories []AttractionCategory
	var total int64

	db := database.GORM_DB.Model(&AttractionCategory{})

	if params.Search != "" {
		term := "%" + params.Search + "%"
		db = db.Where("name ILIKE ? OR code ILIKE ?", term, term)
	}

	db.Count(&total)

	result := db.
		Order(params.OrderClause(attractionCategorySortColumns, "created_at")).
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&categories)

	return categories, total, result.Error
}

func FindByUUID(uuid uuid.UUID) (AttractionCategory, error) {
	var category AttractionCategory
	result := database.GORM_DB.First(&category, "uuid = ?", uuid)
	return category, result.Error
}

func Insert(category *AttractionCategory) error {
	return database.GORM_DB.Create(category).Error
}

func Save(category *AttractionCategory) error {
	return database.GORM_DB.Save(category).Error
}

func Remove(uuid uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", uuid).Delete(&AttractionCategory{}).Error
}

func FindByUUIDs(uuids []uuid.UUID) ([]AttractionCategory, error) {
	var categories []AttractionCategory
	result := database.GORM_DB.Where("uuid IN ?", uuids).Find(&categories)
	return categories, result.Error
}
