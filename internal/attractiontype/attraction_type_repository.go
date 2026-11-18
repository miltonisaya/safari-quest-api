package attractionType

import (
	"safari-quest-api/database"
	"safari-quest-api/pkg/pagination"

	"github.com/google/uuid"
)

var attractionTypeSortColumns = map[string]string{
	"name":       "name",
	"code":       "code",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

func FindAll(params pagination.Params) ([]AttractionType, int64, error) {
	var geographyType []AttractionType
	var total int64

	db := database.GORM_DB.Model(&AttractionType{})

	if params.Search != "" {
		term := "%" + params.Search + "%"
		db = db.Where("name ILIKE ? OR code ILIKE ?", term, term)
	}

	db.Count(&total)

	result := db.
		Order(params.OrderClause(attractionTypeSortColumns, "created_at")).
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&geographyType)

	return geographyType, total, result.Error
}

func FindByUUID(uuid uuid.UUID) (AttractionType, error) {
	var geographyType AttractionType
	result := database.GORM_DB.First(&geographyType, "uuid = ?", uuid)
	return geographyType, result.Error
}

func Insert(geographyType *AttractionType) error {
	return database.GORM_DB.Create(geographyType).Error
}

func Save(geographyType *AttractionType) error {
	return database.GORM_DB.Save(geographyType).Error
}

func Remove(uuid uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", uuid).Delete(&AttractionType{}).Error
}

func FindByUUIDs(uuids []uuid.UUID) ([]AttractionType, error) {
	var geographyType []AttractionType
	result := database.GORM_DB.Where("uuid IN ?", uuids).Find(&geographyType)
	return geographyType, result.Error
}
