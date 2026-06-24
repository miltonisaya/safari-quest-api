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

func RoleFindByID(id uuid.UUID) (models.Role, error) {
	var role models.Role
	result := database.GORM_DB.First(&role, "id = ?", id)
	return role, result.Error
}

func RoleCreate(role *models.Role) error {
	return database.GORM_DB.Create(role).Error
}

func RoleUpdate(role *models.Role) error {
	return database.GORM_DB.Save(role).Error
}

func RoleDelete(id uuid.UUID) error {
	return database.GORM_DB.Delete(&models.Role{}, "id = ?", id).Error
}
