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

func UserFindByEmail(email string) (models.User, error) {
	var user models.User
	result := database.GORM_DB.Preload("Roles").Where("email = ?", email).First(&user)
	return user, result.Error
}

func UserCreate(user *models.User, roles []models.Role) error {
	return database.GORM_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Roles.*").Create(user).Error; err != nil {
			return err
		}
		return tx.Session(&gorm.Session{FullSaveAssociations: false}).
			Model(user).Association("Roles").Replace(roles)
	})
}

func UserUpdate(user *models.User, roles []models.Role) error {
	return database.GORM_DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(user).Error; err != nil {
			return err
		}
		return tx.Session(&gorm.Session{FullSaveAssociations: false}).
			Model(user).Association("Roles").Replace(roles)
	})
}

func UserDelete(uuid uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", uuid).Delete(&models.User{}).Error
}

// UserHasAuthority checks whether the user identified by userUUID holds any
// role that grants the given authority code. It does this in a single JOIN
// query across users → user_roles → roles → role_authorities → authorities
// rather than preloading the full object graph, keeping the hot-path check fast.
func UserHasAuthority(userUUID uuid.UUID, authorityCode string) (bool, error) {
	var count int64
	result := database.GORM_DB.
		Table("users u").
		Joins("JOIN user_roles ur ON ur.user_id = u.id").
		Joins("JOIN roles r ON r.id = ur.role_id").
		Joins("JOIN role_authorities ra ON ra.role_id = r.id").
		Joins("JOIN authorities a ON a.id = ra.authority_id").
		Where("u.uuid = ? AND a.code = ? AND u.deleted_at IS NULL AND r.deleted_at IS NULL AND a.deleted_at IS NULL", userUUID, authorityCode).
		Count(&count)
	return count > 0, result.Error
}
