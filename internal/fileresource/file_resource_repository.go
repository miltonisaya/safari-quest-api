package fileresource

import (
	"safari-quest-api/database"
	"safari-quest-api/pkg/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var sortColumns = map[string]string{
	"name":          "file_resources.name",
	"code":          "file_resources.code",
	"date_uploaded": "file_resources.date_uploaded",
	"created_at":    "file_resources.created_at",
	"updated_at":    "file_resources.updated_at",
}

func withRelations(db *gorm.DB) *gorm.DB {
	return db.Preload("FileResourceType").Preload("Operator")
}

func FindAll(params pagination.Params, operatorIDs []uint, unrestricted bool) ([]FileResource, int64, error) {
	var items []FileResource
	var total int64

	db := database.GORM_DB.Model(&FileResource{})
	if !unrestricted {
		if len(operatorIDs) == 0 {
			db = db.Where("1 = 0")
		} else {
			db = db.Where("file_resources.operator_id IN ?", operatorIDs)
		}
	}
	if params.Search != "" {
		term := "%" + params.Search + "%"
		db = db.Where("file_resources.name ILIKE ? OR file_resources.code ILIKE ?", term, term)
	}
	db.Count(&total)

	result := withRelations(db).
		Order(params.OrderClause(sortColumns, "file_resources.created_at")).
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&items)
	return items, total, result.Error
}

func FindByUUID(id uuid.UUID) (FileResource, error) {
	var item FileResource
	result := withRelations(database.GORM_DB).First(&item, "uuid = ?", id)
	return item, result.Error
}

func FindByUUIDs(ids []uuid.UUID) ([]FileResource, error) {
	var items []FileResource
	if len(ids) == 0 {
		return items, nil
	}
	result := withRelations(database.GORM_DB).Where("uuid IN ?", ids).Find(&items)
	return items, result.Error
}

func Insert(item *FileResource) error {
	return database.GORM_DB.Omit("FileResourceType", "Operator").Create(item).Error
}

func Save(item *FileResource) error {
	return database.GORM_DB.Omit("FileResourceType", "Operator").Save(item).Error
}

// Remove soft-deletes the row. The file on disk is kept.
// shortcut: the stored file is not removed, delete it from UPLOAD_DIR when disk use matters.
func Remove(id uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", id).Delete(&FileResource{}).Error
}
