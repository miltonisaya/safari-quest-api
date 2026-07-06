package repositories

import (
	"safari-quest-api/database"
	"safari-quest-api/models"
	"safari-quest-api/pkg/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var menuSortColumns = map[string]string{
	"name":       "name",
	"sort_order": "sort_order",
	"route":      "route",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

func MenuFindAll(params pagination.Params) ([]models.Menu, int64, error) {
	var menus []models.Menu
	var total int64

	db := database.GORM_DB.Model(&models.Menu{}).Where("parent_id IS NULL")

	if params.Search != "" {
		term := "%" + params.Search + "%"
		db = db.Where("name ILIKE ?", term)
	}

	db.Count(&total)

	result := db.
		Preload("Authority").
		Preload("Children", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Children.Authority").
		Order(params.OrderClause(menuSortColumns, "sort_order")).
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&menus)

	return menus, total, result.Error
}

func MenuFindByUUID(uuid uuid.UUID) (models.Menu, error) {
	var menu models.Menu
	result := database.GORM_DB.
		Preload("Authority").
		Preload("Parent").
		Preload("Children", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Children.Authority").
		First(&menu, "uuid = ?", uuid)
	return menu, result.Error
}

func MenuFindParentByUUID(uuid uuid.UUID) (models.Menu, error) {
	var menu models.Menu
	result := database.GORM_DB.Select("id, uuid").First(&menu, "uuid = ?", uuid)
	return menu, result.Error
}

func MenuCreate(menu *models.Menu) error {
	return database.GORM_DB.Create(menu).Error
}

func MenuUpdate(menu *models.Menu) error {
	return database.GORM_DB.Save(menu).Error
}

func MenuDelete(uuid uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", uuid).Delete(&models.Menu{}).Error
}

func MenuFindByUUIDs(uuids []uuid.UUID) ([]models.Menu, error) {
	var menus []models.Menu
	result := database.GORM_DB.Where("uuid IN ?", uuids).Find(&menus)
	return menus, result.Error
}
