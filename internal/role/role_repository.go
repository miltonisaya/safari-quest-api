package role

import (
	"safari-quest-api/database"
	"safari-quest-api/pkg/pagination"

	"github.com/google/uuid"
)

var roleSortColumns = map[string]string{
	"name":       "name",
	"code":       "code",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

func FindAll(params pagination.Params) ([]Role, int64, error) {
	var roles []Role
	var total int64

	db := database.GORM_DB.Model(&Role{})

	if params.Search != "" {
		term := "%" + params.Search + "%"
		db = db.Where("name ILIKE ? OR code ILIKE ?", term, term)
	}

	db.Count(&total)

	result := db.
		Order(params.OrderClause(roleSortColumns, "created_at")).
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&roles)

	return roles, total, result.Error
}

func FindByUUID(uuid uuid.UUID) (Role, error) {
	var role Role
	result := database.GORM_DB.First(&role, "uuid = ?", uuid)
	return role, result.Error
}

func FindByUUIDWithAuthorities(uuid uuid.UUID) (Role, error) {
	var role Role
	result := database.GORM_DB.Preload("Authorities").First(&role, "uuid = ?", uuid)
	return role, result.Error
}

func Insert(role *Role) error {
	return database.GORM_DB.Create(role).Error
}

func Save(role *Role) error {
	return database.GORM_DB.Save(role).Error
}

func Remove(uuid uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", uuid).Delete(&Role{}).Error
}

func FindByUUIDs(uuids []uuid.UUID) ([]Role, error) {
	var roles []Role
	result := database.GORM_DB.Where("uuid IN ?", uuids).Find(&roles)
	return roles, result.Error
}
