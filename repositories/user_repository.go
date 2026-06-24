package repositories

import (
	"safari-quest-api/database"
	"safari-quest-api/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func UserFindAll() ([]models.User, error) {
	var users []models.User
	result := database.GORM_DB.Preload("Roles").Find(&users)
	return users, result.Error
}

func UserFindByUUID(uuid uuid.UUID) (models.User, error) {
	var user models.User
	result := database.GORM_DB.Preload("Roles").First(&user, "uuid = ?", uuid)
	return user, result.Error
}

func UserCreate(user *models.User) error {
	return database.GORM_DB.Create(user).Error
}

func UserUpdate(user *models.User, roles []models.Role) error {
	return database.GORM_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(user).Error; err != nil {
			return err
		}
		return tx.Model(user).Association("Roles").Replace(roles)
	})
}

func UserDelete(uuid uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", uuid).Delete(&models.User{}).Error
}
