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

func GeographyFindAll(params pagination.Params, parentUUID *uuid.UUID) ([]models.Geography, int64, error) {
	var geographies []models.Geography
	var total int64

	db := database.GORM_DB.Model(&models.Geography{})

	if parentUUID != nil {
		// Explicit parent filter — return only direct children of that parent.
		parent, err := GeographyFindParentByUUID(*parentUUID)
		if err != nil {
			return nil, 0, err
		}
		db = db.Where("parent_id = ?", parent.ID)
	} else if params.Search == "" {
		// No parent and no search — return top-level geographies only.
		db = db.Where("parent_id IS NULL")
	}
	// When a search term is present with no parent filter, search across all levels.

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
	result := database.GORM_DB.
		Preload("GeographyType").
		Preload("Parent").
		Preload("Children").
		Preload("Children.GeographyType").
		First(&geography, "uuid = ?", uuid)
	return geography, result.Error
}

func GeographyFindParentByUUID(uuid uuid.UUID) (models.Geography, error) {
	var geography models.Geography
	result := database.GORM_DB.Select("id, uuid").First(&geography, "uuid = ?", uuid)
	return geography, result.Error
}

func GeographyCreate(geography *models.Geography) error {
	return database.GORM_DB.Create(geography).Error
}

func GeographyUpdate(geography *models.Geography) error {
	updates := map[string]interface{}{
		"name":              geography.Name,
		"code":              geography.Code,
		"geography_type_id": geography.GeographyTypeID,
		"parent_id":         nil,
	}
	if geography.ParentID != nil {
		updates["parent_id"] = *geography.ParentID
	}
	// Use a clean model with only the primary key to prevent GORM from trying
	// to upsert preloaded associations (Parent, Children) before the update.
	return database.GORM_DB.Model(&models.Geography{ID: geography.ID}).Updates(updates).Error
}

func GeographyDelete(uuid uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", uuid).Delete(&models.Geography{}).Error
}

func GeographyFindByUUIDs(uuids []uuid.UUID) ([]models.Geography, error) {
	var geographies []models.Geography
	result := database.GORM_DB.Where("uuid IN ?", uuids).Find(&geographies)
	return geographies, result.Error
}
