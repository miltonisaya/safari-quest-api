package setting

import (
	"safari-quest-api/database"

	"gorm.io/gorm/clause"
)

func findAll() ([]Setting, error) {
	var rows []Setting
	err := database.GORM_DB.Find(&rows).Error
	return rows, err
}

func findValue(key string) (string, bool) {
	var row Setting
	if err := database.GORM_DB.Where("key = ?", key).Take(&row).Error; err != nil {
		return "", false
	}
	return row.Value, true
}

// upsert writes the values in one transaction, inserting or replacing each key.
func upsert(rows []Setting) error {
	return database.GORM_DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_by", "updated_at"}),
	}).Create(&rows).Error
}
