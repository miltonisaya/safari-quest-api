package repositories

import (
	"safari-quest-api/database"
	"safari-quest-api/models"
	"safari-quest-api/pkg/pagination"

	"github.com/google/uuid"
)

var geographyTypeSortColumns = map[string]string{
	"name":       "name",
	"code":       "code",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

func GeographyTypeFindAll(params pagination.Params) ([]models.GeographyType, int64, error) {
	var geographyType []models.GeographyType
	var total int64

	db := database.GORM_DB.Model(&models.GeographyType{})

	if params.Search != "" {
		term := "%" + params.Search + "%"
		db = db.Where("name ILIKE ? OR code ILIKE ?", term, term)
	}

	db.Count(&total)

	result := db.
		Order(params.OrderClause(geographySortColumns, "created_at")).
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&geographyType)

	return geographyType, total, result.Error
}

func GeographyTypeFindByUUID(uuid uuid.UUID) (models.GeographyType, error) {
	var geographyType models.GeographyType
	result := database.GORM_DB.First(&geographyType, "uuid = ?", uuid)
	return geographyType, result.Error
}

func GeographyTypeCreate(geographyType *models.GeographyType) error {
	return database.GORM_DB.Create(geographyType).Error
}

func GeographyTypeUpdate(geographyType *models.GeographyType) error {
	return database.GORM_DB.Save(geographyType).Error
}

func GeographyTypeDelete(uuid uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", uuid).Delete(&models.GeographyType{}).Error
}

func GeographyTypeFindByUUIDs(uuids []uuid.UUID) ([]models.GeographyType, error) {
	var geographyType []models.GeographyType
	result := database.GORM_DB.Where("uuid IN ?", uuids).Find(&geographyType)
	return geographyType, result.Error
}
