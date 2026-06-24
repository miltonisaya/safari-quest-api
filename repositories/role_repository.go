package repositories

import (
	"safari-quest-api/database"
	"safari-quest-api/models"

	"github.com/google/uuid"
)

func RoleFindAll() ([]models.Role, error) {
	var roles []models.Role
	result := database.GORM_DB.Find(&roles)
	return roles, result.Error
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
