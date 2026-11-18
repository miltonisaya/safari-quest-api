package authority

import (
	"safari-quest-api/database"

	"github.com/google/uuid"
)

func FindAll() ([]Authority, error) {
	var authorities []Authority
	result := database.GORM_DB.Find(&authorities)
	return authorities, result.Error
}

func FindByUUID(uuid uuid.UUID) (Authority, error) {
	var authority Authority
	result := database.GORM_DB.First(&authority, "uuid = ?", uuid)
	return authority, result.Error
}

func Insert(authority *Authority) error {
	return database.GORM_DB.Create(authority).Error
}

func Save(authority *Authority) error {
	return database.GORM_DB.Save(authority).Error
}

func Remove(uuid uuid.UUID) error {
	return database.GORM_DB.Where("uuid = ?", uuid).Delete(&Authority{}).Error
}

func FindByUUIDs(uuids []uuid.UUID) ([]Authority, error) {
	var authorities []Authority
	result := database.GORM_DB.Where("uuid IN ?", uuids).Find(&authorities)
	return authorities, result.Error
}
