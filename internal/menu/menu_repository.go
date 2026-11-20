package menu

import (
	"safari-quest-api/database"
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

func orderedChildren(db *gorm.DB) *gorm.DB {
	return db.Order("sort_order ASC")
}

func FindAll(params pagination.Params) ([]Menu, int64, error) {
	var menus []Menu
	var total int64

	db := database.GORM_DB.Model(&Menu{}).Where("parent_id IS NULL")

	if params.Search != "" {
		term := "%" + params.Search + "%"
		db = db.Where("name ILIKE ?", term)
	}

	db.Count(&total)

	result := db.
		Preload("Authority").
		Preload("Children", orderedChildren).
		Preload("Children.Authority").
		Preload("Children.Children", orderedChildren).
		Preload("Children.Children.Authority").
		Order(params.OrderClause(menuSortColumns, "sort_order")).
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&menus)

	return menus, total, result.Error
}

func FindByUUID(uuid uuid.UUID) (Menu, error) {
	var menu Menu
	result := database.GORM_DB.
		Preload("Authority").
		Preload("Parent").
		Preload("Children", orderedChildren).
		Preload("Children.Authority").
		Preload("Children.Children", orderedChildren).
		Preload("Children.Children.Authority").
		First(&menu, "uuid = ?", uuid)
	return menu, result.Error
}

func FindParentByUUID(uuid uuid.UUID) (Menu, error) {
	var menu Menu
	result := database.GORM_DB.Select("id, uuid").First(&menu, "uuid = ?", uuid)
	return menu, result.Error
}

func Insert(menu *Menu) error {
	return database.GORM_DB.Create(menu).Error
}

func Save(menu *Menu) error {
	return database.GORM_DB.Save(menu).Error
}

func Remove(uuid uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", uuid).Delete(&Menu{}).Error
}

func FindByUUIDs(uuids []uuid.UUID) ([]Menu, error) {
	var menus []Menu
	result := database.GORM_DB.Where("uuid IN ?", uuids).Find(&menus)
	return menus, result.Error
}
