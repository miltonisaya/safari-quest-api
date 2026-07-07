package repositories

import (
	"safari-quest-api/database"
	"safari-quest-api/models"
	"safari-quest-api/pkg/pagination"

	"github.com/google/uuid"
)

var geographySortColumns = map[string]string{
	"name":       "name",
	"code":       "code",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

func GeographyFindAll(params pagination.Params) ([]models.Geography, int64, error) {
	var geographies []models.Geography
	var total int64

	db := database.GORM_DB.Model(&models.Geography{})

	if params.Search != "" {
		term := "%" + params.Search + "%"
		db = db.Where("name ILIKE ? OR code ILIKE ?", term, term)
	}

	db.Count(&total)

	result := db.
		Preload("GeographyType").
		Order(params.OrderClause(geographySortColumns, "created_at")).
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&geographies)

	return geographies, total, result.Error
}

func GeographyFindByUUID(uuid uuid.UUID) (models.Geography, error) {
	var geography models.Geography
	result := database.GORM_DB.Preload("GeographyType").First(&geography, "uuid = ?", uuid)
	return geography, result.Error
}

func GeographyCreate(geography *models.Geography) error {
	return database.GORM_DB.Create(geography).Error
}

func GeographyUpdate(geography *models.Geography) error {
	return database.GORM_DB.Save(geography).Error
}

func GeographyDelete(uuid uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", uuid).Delete(&models.Geography{}).Error
}

func GeographyFindByUUIDs(uuids []uuid.UUID) ([]models.Geography, error) {
	var geographies []models.Geography
	result := database.GORM_DB.Where("uuid IN ?", uuids).Find(&geographies)
	return geographies, result.Error
}
