package fileresourcetype

import (
	"safari-quest-api/database"
	"safari-quest-api/pkg/pagination"

	"github.com/google/uuid"
)

var sortColumns = map[string]string{
	"name":       "name",
	"code":       "code",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

func FindAll(params pagination.Params) ([]FileResourceType, int64, error) {
	var items []FileResourceType
	var total int64

	db := database.GORM_DB.Model(&FileResourceType{})
	if params.Search != "" {
		term := "%" + params.Search + "%"
		db = db.Where("name ILIKE ? OR code ILIKE ?", term, term)
	}
	db.Count(&total)

	result := db.
		Order(params.OrderClause(sortColumns, "created_at")).
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&items)
	return items, total, result.Error
}

func FindByUUID(id uuid.UUID) (FileResourceType, error) {
	var item FileResourceType
	result := database.GORM_DB.First(&item, "uuid = ?", id)
	return item, result.Error
}

func FindByUUIDs(ids []uuid.UUID) ([]FileResourceType, error) {
	var items []FileResourceType
	result := database.GORM_DB.Where("uuid IN ?", ids).Find(&items)
	return items, result.Error
}

func Insert(item *FileResourceType) error {
	return database.GORM_DB.Create(item).Error
}

func Save(item *FileResourceType) error {
	return database.GORM_DB.Save(item).Error
}

func Remove(id uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", id).Delete(&FileResourceType{}).Error
}
