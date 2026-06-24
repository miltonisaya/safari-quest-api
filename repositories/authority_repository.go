package repositories

import (
	"safari-quest-api/database"
	"safari-quest-api/models"

	"github.com/google/uuid"
)

func AuthorityFindAll() ([]models.Authority, error) {
	var authorities []models.Authority
	result := database.GORM_DB.Find(&authorities)
	return authorities, result.Error
}

func AuthorityFindByUUID(uuid uuid.UUID) (models.Authority, error) {
	var authority models.Authority
	result := database.GORM_DB.First(&authority, "uuid = ?", uuid)
	return authority, result.Error
}

func AuthorityCreate(authority *models.Authority) error {
	return database.GORM_DB.Create(authority).Error
}

func AuthorityUpdate(authority *models.Authority) error {
	return database.GORM_DB.Save(authority).Error
}

func AuthorityDelete(uuid uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", uuid).Delete(&models.Authority{}).Error
}
