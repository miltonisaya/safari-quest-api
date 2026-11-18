package geographytype

import (
	"safari-quest-api/database"
	"safari-quest-api/pkg/pagination"

	"github.com/google/uuid"
)

var geographyTypeSortColumns = map[string]string{
	"name":       "name",
	"code":       "code",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

func FindAll(params pagination.Params) ([]GeographyType, int64, error) {
	var geographyType []GeographyType
	var total int64

	db := database.GORM_DB.Model(&GeographyType{})

	if params.Search != "" {
		term := "%" + params.Search + "%"
		db = db.Where("name ILIKE ? OR code ILIKE ?", term, term)
	}

	db.Count(&total)

	result := db.
		Order(params.OrderClause(geographyTypeSortColumns, "created_at")).
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&geographyType)

	return geographyType, total, result.Error
}

func FindByUUID(uuid uuid.UUID) (GeographyType, error) {
	var geographyType GeographyType
	result := database.GORM_DB.First(&geographyType, "uuid = ?", uuid)
	return geographyType, result.Error
}

func Insert(geographyType *GeographyType) error {
	return database.GORM_DB.Create(geographyType).Error
}

func Save(geographyType *GeographyType) error {
	return database.GORM_DB.Save(geographyType).Error
}

func Remove(uuid uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", uuid).Delete(&GeographyType{}).Error
}

func FindByUUIDs(uuids []uuid.UUID) ([]GeographyType, error) {
	var geographyType []GeographyType
	result := database.GORM_DB.Where("uuid IN ?", uuids).Find(&geographyType)
	return geographyType, result.Error
}
