package repositories

import (
	"safari-quest-api/database"
	"safari-quest-api/models"
	"safari-quest-api/pkg/pagination"

	"github.com/google/uuid"
)

var roleSortColumns = map[string]string{
	"name":       "name",
	"code":       "code",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

func RoleFindAll(params pagination.Params) ([]models.Role, int64, error) {
	var roles []models.Role
	var total int64

	db := database.GORM_DB.Model(&models.Role{})

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

func RoleFindByUUID(uuid uuid.UUID) (models.Role, error) {
	var role models.Role
	result := database.GORM_DB.First(&role, "uuid = ?", uuid)
	return role, result.Error
}

func RoleCreate(role *models.Role) error {
	return database.GORM_DB.Create(role).Error
}

func RoleUpdate(role *models.Role) error {
	return database.GORM_DB.Save(role).Error
}

func RoleDelete(uuid uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", uuid).Delete(&models.Role{}).Error
}

func RoleFindByUUIDs(uuids []uuid.UUID) ([]models.Role, error) {
	var roles []models.Role
	result := database.GORM_DB.Where("uuid IN ?", uuids).Find(&roles)
	return roles, result.Error
}
