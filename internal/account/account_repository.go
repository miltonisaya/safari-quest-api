package account

import (
	"safari-quest-api/database"
	"safari-quest-api/internal/user"
)

// saveProfile updates only the profile columns, never email, roles or status.
func saveProfile(id uint, fields map[string]any) error {
	return database.GORM_DB.Model(&user.User{}).Where("id = ?", id).Updates(fields).Error
}

func savePassword(id uint, hashed string) error {
	return database.GORM_DB.Model(&user.User{}).Where("id = ?", id).Update("password", hashed).Error
}
